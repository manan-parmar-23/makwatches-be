package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/middleware"
)

// PaymentHandler provides endpoints for initiating payments (Razorpay order creation)
type PaymentHandler struct {
	DB  *database.DBClient
	Cfg *config.Config
}

func NewPaymentHandler(db *database.DBClient, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{DB: db, Cfg: cfg}
}

// cartTotalINR computes the current cart total for a user
func (h *PaymentHandler) cartTotalINR(userID any) (float64, error) {
	ctx := context.Background()
	cartCol := h.DB.Collections().CartItems
	prodCol := h.DB.Collections().Products
	cursor, err := cartCol.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	type itemRow struct {
		ProductID interface{} `bson:"product_id"`
		Quantity  int         `bson:"quantity"`
	}
	var rows []itemRow
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, err
	}
	total := 0.0
	for _, r := range rows {
		var p struct {
			Price float64 `bson:"price"`
			Stock int     `bson:"stock"`
		}
		if err := prodCol.FindOne(ctx, bson.M{"_id": r.ProductID}).Decode(&p); err != nil {
			return 0, err
		}
		if p.Stock < r.Quantity {
			return 0, fmt.Errorf("insufficient stock for a product")
		}
		total += p.Price * float64(r.Quantity)
	}
	return total, nil
}

// CreateRazorpayOrder creates a Razorpay order from cart total
func (h *PaymentHandler) CreateRazorpayOrder(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	if h.Cfg.RazorpayKey == "" || h.Cfg.RazorpaySecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "Payment gateway not configured"})
	}
	total, err := h.cartTotalINR(user.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	if total <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Cart empty"})
	}

	amountPaise := int64(total * 100)
	rnd := make([]byte, 6)
	rand.Read(rnd)
	receipt := fmt.Sprintf("rcpt_%s", hex.EncodeToString(rnd))

	payload := map[string]any{"amount": amountPaise, "currency": "INR", "receipt": receipt, "payment_capture": 1}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(h.Cfg.RazorpayKey, h.Cfg.RazorpaySecret)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Failed to create payment order", "error": err.Error()})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"success": false, "message": "Gateway error", "raw": string(body)})
	}

	return c.JSON(fiber.Map{"success": true, "key": h.Cfg.RazorpayKey, "amount": amountPaise, "currency": "INR", "data": json.RawMessage(body)})
}

// RazorpayWebhook validates webhook signatures from Razorpay
// Set the endpoint URL in Razorpay dashboard and use Cfg.RazorpayWebhookSecret
func (h *PaymentHandler) RazorpayWebhook(c *fiber.Ctx) error {
	if h.Cfg.RazorpayWebhookSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "Webhook secret not configured"})
	}

	sig := c.Get("X-Razorpay-Signature")
	if sig == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Missing signature"})
	}

	body := c.Body()
	mac := hmac.New(sha256.New, []byte(h.Cfg.RazorpayWebhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid webhook signature"})
	}

	// Parse event (optional minimal handling)
	var evt map[string]any
	if err := json.Unmarshal(body, &evt); err == nil {
		// You can extend: update order/payment status based on event
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
