package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentInfo represents payment information
type PaymentInfo struct {
	Method            string `json:"method" bson:"method"` // "razorpay", "card", "cod", etc.
	CardNumber        string `json:"cardNumber,omitempty" bson:"card_number,omitempty"`
	ExpiryDate        string `json:"expiryDate,omitempty" bson:"expiry_date,omitempty"`
	CVV               string `json:"cvv,omitempty" bson:"-"` // Never store CVV
	RazorpayOrderID   string `json:"razorpayOrderId,omitempty" bson:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string `json:"razorpayPaymentId,omitempty" bson:"razorpay_payment_id,omitempty"`
	RazorpaySignature string `json:"razorpaySignature,omitempty" bson:"razorpay_signature,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID   primitive.ObjectID `json:"productId" bson:"product_id"`
	ProductName string             `json:"productName" bson:"product_name"`
	Price       float64            `json:"price" bson:"price"`
	Size        string             `json:"size,omitempty" bson:"size,omitempty"`
	Quantity    int                `json:"quantity" bson:"quantity"`
	Subtotal    float64            `json:"subtotal" bson:"subtotal"`
}

// Order represents a user order
type Order struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"` // <-- ensure json:"id"
	UserID          primitive.ObjectID `json:"userId" bson:"user_id"`   // <-- ensure json:"userId"
	Items           []OrderItem        `json:"items" bson:"items"`
	Total           float64            `json:"total" bson:"total"`
	Status          string             `json:"status" bson:"status"`
	PaymentStatus   string             `json:"paymentStatus" bson:"payment_status"`
	ShippingAddress Address            `json:"shippingAddress" bson:"shipping_address"`
	PaymentInfo     PaymentInfo        `json:"paymentInfo" bson:"payment_info"`
	CreatedAt       time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CheckoutRequest represents the data required for placing an order
type CheckoutRequest struct {
	UserID          string      `json:"userId" validate:"required"`
	ShippingAddress Address     `json:"shippingAddress" validate:"required"`
	PaymentInfo     PaymentInfo `json:"paymentInfo" validate:"required"`
	ClientTotal     *float64    `json:"clientTotal,omitempty" bson:"-"`
}
