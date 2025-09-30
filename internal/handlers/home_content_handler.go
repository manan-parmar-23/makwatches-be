package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/models"
)

const (
	heroSlidesCollectionName         = "hero_slides"
	categoryCardsCollectionName      = "home_category_cards"
	collectionFeaturesCollectionName = "home_collection_features"
	techCardsCollectionName          = "home_tech_cards"
	techHighlightCollectionName      = "home_tech_highlights"
	homeContentCacheKey              = "home_content"
)

// HomeContentHandler manages curated landing page data.
type HomeContentHandler struct {
	DB *database.DBClient
}

// NewHomeContentHandler wires a handler with the provided DB client.
func NewHomeContentHandler(db *database.DBClient) *HomeContentHandler {
	return &HomeContentHandler{DB: db}
}

// GetHomeContent returns aggregated landing page content for the storefront.
func (h *HomeContentHandler) GetHomeContent(c *fiber.Ctx) error {
	ctx := c.Context()

	var cached models.HomeContent
	if err := h.DB.CacheGet(ctx, homeContentCacheKey, &cached); err == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Home content retrieved from cache",
			"data":    cached,
		})
	}

	heroSlides, err := h.fetchHeroSlides(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch hero slides")
	}

	categories, err := h.fetchCategoryCards(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch category cards")
	}

	collections, err := h.fetchCollectionFeatures(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch collection features")
	}

	techCards, err := h.fetchTechCards(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch tech showcase cards")
	}

	highlight, err := h.fetchTechHighlight(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch tech highlight")
	}

	payload := models.HomeContent{
		HeroSlides:  heroSlides,
		Categories:  categories,
		Collections: collections,
		TechCards:   techCards,
		Highlight:   highlight,
	}

	// Cache for five minutes to avoid excessive DB hits while remaining responsive to updates.
	_ = h.DB.CacheSet(ctx, homeContentCacheKey, payload, 5*time.Minute)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Home content retrieved successfully",
		"data":    payload,
	})
}

// ============ Hero Slides CRUD ============

// ListHeroSlides returns all hero slides for admin management.
func (h *HomeContentHandler) ListHeroSlides(c *fiber.Ctx) error {
	ctx := c.Context()
	slides, err := h.fetchHeroSlides(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch hero slides")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Hero slides retrieved successfully",
		"data":    slides,
	})
}

// CreateHeroSlide inserts a new hero slide document.
func (h *HomeContentHandler) CreateHeroSlide(c *fiber.Ctx) error {
	ctx := c.Context()
	var payload models.HeroSlide
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateHeroSlide(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	coll := h.DB.MongoDB.Collection(heroSlidesCollectionName)
	now := time.Now().UTC()
	payload.ID = primitive.NilObjectID
	payload.CreatedAt = now
	payload.UpdatedAt = now
	if payload.Position <= 0 {
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err == nil {
			payload.Position = int(count) + 1
		} else {
			payload.Position = 1
		}
	}

	res, err := coll.InsertOne(ctx, payload)
	if err != nil {
		return fiberError(c, err, "Failed to create hero slide")
	}

	if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
		payload.ID = insertedID
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Hero slide created",
		"data":    payload,
	})
}

// UpdateHeroSlide updates an existing hero slide document.
func (h *HomeContentHandler) UpdateHeroSlide(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid hero slide id", err)
	}

	var payload models.HeroSlide
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateHeroSlide(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	update := bson.M{
		"title":       payload.Title,
		"subtitle":    payload.Subtitle,
		"price":       payload.Price,
		"description": payload.Description,
		"image":       payload.Image,
		"features":    payload.Features,
		"gradient":    payload.Gradient,
		"glowColor":   payload.GlowColor,
		"updatedAt":   time.Now().UTC(),
	}
	if payload.Position > 0 {
		update["position"] = payload.Position
	}

	coll := h.DB.MongoDB.Collection(heroSlidesCollectionName)
	result, err := coll.UpdateByID(ctx, objectID, bson.M{"$set": update})
	if err != nil {
		return fiberError(c, err, "Failed to update hero slide")
	}
	if result.MatchedCount == 0 {
		return fiberNotFound(c, "Hero slide not found")
	}

	var updated models.HeroSlide
	if err := coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updated); err != nil {
		return fiberError(c, err, "Failed to load updated hero slide")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Hero slide updated",
		"data":    updated,
	})
}

// DeleteHeroSlide removes a hero slide by id.
func (h *HomeContentHandler) DeleteHeroSlide(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid hero slide id", err)
	}

	coll := h.DB.MongoDB.Collection(heroSlidesCollectionName)
	res, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fiberError(c, err, "Failed to delete hero slide")
	}
	if res.DeletedCount == 0 {
		return fiberNotFound(c, "Hero slide not found")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Hero slide deleted",
	})
}

// ============ Category Cards CRUD ============

func (h *HomeContentHandler) ListCategoryCards(c *fiber.Ctx) error {
	ctx := c.Context()
	cards, err := h.fetchCategoryCards(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch category cards")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Category cards retrieved successfully",
		"data":    cards,
	})
}

func (h *HomeContentHandler) CreateCategoryCard(c *fiber.Ctx) error {
	ctx := c.Context()
	var payload models.HomeCategoryCard
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateCategoryCard(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	coll := h.DB.MongoDB.Collection(categoryCardsCollectionName)
	now := time.Now().UTC()
	payload.ID = primitive.NilObjectID
	payload.CreatedAt = now
	payload.UpdatedAt = now
	if payload.Position <= 0 {
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err == nil {
			payload.Position = int(count) + 1
		} else {
			payload.Position = 1
		}
	}

	res, err := coll.InsertOne(ctx, payload)
	if err != nil {
		return fiberError(c, err, "Failed to create category card")
	}
	if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
		payload.ID = insertedID
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Category card created",
		"data":    payload,
	})
}

func (h *HomeContentHandler) UpdateCategoryCard(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid category card id", err)
	}

	var payload models.HomeCategoryCard
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateCategoryCard(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	update := bson.M{
		"title":      payload.Title,
		"subtitle":   payload.Subtitle,
		"href":       payload.Href,
		"image":      payload.Image,
		"bgGradient": payload.BgGradient,
		"updatedAt":  time.Now().UTC(),
	}
	if payload.Position > 0 {
		update["position"] = payload.Position
	}

	coll := h.DB.MongoDB.Collection(categoryCardsCollectionName)
	res, err := coll.UpdateByID(ctx, objectID, bson.M{"$set": update})
	if err != nil {
		return fiberError(c, err, "Failed to update category card")
	}
	if res.MatchedCount == 0 {
		return fiberNotFound(c, "Category card not found")
	}

	var updated models.HomeCategoryCard
	if err := coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updated); err != nil {
		return fiberError(c, err, "Failed to load updated category card")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Category card updated",
		"data":    updated,
	})
}

func (h *HomeContentHandler) DeleteCategoryCard(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid category card id", err)
	}

	coll := h.DB.MongoDB.Collection(categoryCardsCollectionName)
	res, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fiberError(c, err, "Failed to delete category card")
	}
	if res.DeletedCount == 0 {
		return fiberNotFound(c, "Category card not found")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Category card deleted",
	})
}

// ============ Collection Features CRUD ============

func (h *HomeContentHandler) ListCollectionFeatures(c *fiber.Ctx) error {
	ctx := c.Context()
	cards, err := h.fetchCollectionFeatures(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch collection features")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection features retrieved successfully",
		"data":    cards,
	})
}

func (h *HomeContentHandler) CreateCollectionFeature(c *fiber.Ctx) error {
	ctx := c.Context()
	var payload models.HomeCollectionFeature
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateCollectionFeature(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	coll := h.DB.MongoDB.Collection(collectionFeaturesCollectionName)
	now := time.Now().UTC()
	payload.ID = primitive.NilObjectID
	payload.CreatedAt = now
	payload.UpdatedAt = now
	if payload.Position <= 0 {
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err == nil {
			payload.Position = int(count) + 1
		} else {
			payload.Position = 1
		}
	}

	res, err := coll.InsertOne(ctx, payload)
	if err != nil {
		return fiberError(c, err, "Failed to create collection feature")
	}
	if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
		payload.ID = insertedID
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Collection feature created",
		"data":    payload,
	})
}

func (h *HomeContentHandler) UpdateCollectionFeature(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid collection feature id", err)
	}

	var payload models.HomeCollectionFeature
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateCollectionFeature(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	update := bson.M{
		"tagline":      payload.Tagline,
		"title":        payload.Title,
		"description":  payload.Description,
		"availability": payload.Availability,
		"ctaLabel":     payload.CtaLabel,
		"ctaHref":      payload.CtaHref,
		"image":        payload.Image,
		"imageAlt":     payload.ImageAlt,
		"layout":       payload.Layout,
		"updatedAt":    time.Now().UTC(),
	}
	if payload.Position > 0 {
		update["position"] = payload.Position
	}

	coll := h.DB.MongoDB.Collection(collectionFeaturesCollectionName)
	res, err := coll.UpdateByID(ctx, objectID, bson.M{"$set": update})
	if err != nil {
		return fiberError(c, err, "Failed to update collection feature")
	}
	if res.MatchedCount == 0 {
		return fiberNotFound(c, "Collection feature not found")
	}

	var updated models.HomeCollectionFeature
	if err := coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updated); err != nil {
		return fiberError(c, err, "Failed to load updated collection feature")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection feature updated",
		"data":    updated,
	})
}

func (h *HomeContentHandler) DeleteCollectionFeature(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid collection feature id", err)
	}

	coll := h.DB.MongoDB.Collection(collectionFeaturesCollectionName)
	res, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fiberError(c, err, "Failed to delete collection feature")
	}
	if res.DeletedCount == 0 {
		return fiberNotFound(c, "Collection feature not found")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection feature deleted",
	})
}

// ============ Tech Showcase Cards & Highlight ============

func (h *HomeContentHandler) ListTechCards(c *fiber.Ctx) error {
	ctx := c.Context()
	cards, err := h.fetchTechCards(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch tech cards")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech cards retrieved successfully",
		"data":    cards,
	})
}

func (h *HomeContentHandler) CreateTechCard(c *fiber.Ctx) error {
	ctx := c.Context()
	var payload models.TechShowcaseCard
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateTechCard(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	coll := h.DB.MongoDB.Collection(techCardsCollectionName)
	now := time.Now().UTC()
	payload.ID = primitive.NilObjectID
	payload.CreatedAt = now
	payload.UpdatedAt = now
	if payload.Position <= 0 {
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err == nil {
			payload.Position = int(count) + 1
		} else {
			payload.Position = 1
		}
	}

	res, err := coll.InsertOne(ctx, payload)
	if err != nil {
		return fiberError(c, err, "Failed to create tech card")
	}
	if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
		payload.ID = insertedID
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tech card created",
		"data":    payload,
	})
}

func (h *HomeContentHandler) UpdateTechCard(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid tech card id", err)
	}

	var payload models.TechShowcaseCard
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateTechCard(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	update := bson.M{
		"title":           payload.Title,
		"subtitle":        payload.Subtitle,
		"image":           payload.Image,
		"backgroundImage": payload.BackgroundImage,
		"rating":          payload.Rating,
		"reviewCount":     payload.ReviewCount,
		"badge":           payload.Badge,
		"color":           payload.Color,
		"updatedAt":       time.Now().UTC(),
	}
	if payload.Position > 0 {
		update["position"] = payload.Position
	}

	coll := h.DB.MongoDB.Collection(techCardsCollectionName)
	res, err := coll.UpdateByID(ctx, objectID, bson.M{"$set": update})
	if err != nil {
		return fiberError(c, err, "Failed to update tech card")
	}
	if res.MatchedCount == 0 {
		return fiberNotFound(c, "Tech card not found")
	}

	var updated models.TechShowcaseCard
	if err := coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updated); err != nil {
		return fiberError(c, err, "Failed to load updated tech card")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech card updated",
		"data":    updated,
	})
}

func (h *HomeContentHandler) DeleteTechCard(c *fiber.Ctx) error {
	ctx := c.Context()
	objectID, err := parseObjectID(c.Params("id"))
	if err != nil {
		return fiberBadRequest(c, "Invalid tech card id", err)
	}

	coll := h.DB.MongoDB.Collection(techCardsCollectionName)
	res, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fiberError(c, err, "Failed to delete tech card")
	}
	if res.DeletedCount == 0 {
		return fiberNotFound(c, "Tech card not found")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech card deleted",
	})
}

// GetTechHighlight returns the current showcase highlight for admin editing.
func (h *HomeContentHandler) GetTechHighlight(c *fiber.Ctx) error {
	ctx := c.Context()
	highlight, err := h.fetchTechHighlight(ctx)
	if err != nil {
		return fiberError(c, err, "Failed to fetch tech highlight")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech highlight retrieved successfully",
		"data":    highlight,
	})
}

// UpsertTechHighlight creates or updates the single tech highlight document.
func (h *HomeContentHandler) UpsertTechHighlight(c *fiber.Ctx) error {
	ctx := c.Context()
	var payload models.TechShowcaseHighlight
	if err := c.BodyParser(&payload); err != nil {
		return fiberBadRequest(c, "Invalid payload", err)
	}
	if err := validateHighlight(&payload); err != nil {
		return fiberBadRequest(c, err.Error(), err)
	}

	coll := h.DB.MongoDB.Collection(techHighlightCollectionName)
	now := time.Now().UTC()

	update := bson.M{
		"value":      payload.Value,
		"title":      payload.Title,
		"subtitle":   payload.Subtitle,
		"accentHex":  payload.AccentHex,
		"background": payload.Background,
		"updatedAt":  now,
	}

	// Upsert to ensure a single document exists.
	opts := options.Update().SetUpsert(true)
	if _, err := coll.UpdateOne(ctx, bson.M{}, bson.M{"$set": update, "$setOnInsert": bson.M{"createdAt": now}}, opts); err != nil {
		return fiberError(c, err, "Failed to upsert tech highlight")
	}

	var highlight models.TechShowcaseHighlight
	if err := coll.FindOne(ctx, bson.M{}).Decode(&highlight); err != nil {
		return fiberError(c, err, "Failed to load tech highlight")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech highlight saved",
		"data":    highlight,
	})
}

// DeleteTechHighlight removes the highlight document (optional cleanup).
func (h *HomeContentHandler) DeleteTechHighlight(c *fiber.Ctx) error {
	ctx := c.Context()
	coll := h.DB.MongoDB.Collection(techHighlightCollectionName)
	res, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fiberError(c, err, "Failed to delete tech highlight")
	}
	if res.DeletedCount == 0 {
		return fiberNotFound(c, "No tech highlight to delete")
	}

	h.clearHomeCache(ctx)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tech highlight deleted",
	})
}

// ============ Helper functions ============

func (h *HomeContentHandler) fetchHeroSlides(ctx context.Context) ([]models.HeroSlide, error) {
	coll := h.DB.MongoDB.Collection(heroSlidesCollectionName)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slides []models.HeroSlide
	if err := cursor.All(ctx, &slides); err != nil {
		return nil, err
	}
	return slides, nil
}

func (h *HomeContentHandler) fetchCategoryCards(ctx context.Context) ([]models.HomeCategoryCard, error) {
	coll := h.DB.MongoDB.Collection(categoryCardsCollectionName)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.HomeCategoryCard
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func (h *HomeContentHandler) fetchCollectionFeatures(ctx context.Context) ([]models.HomeCollectionFeature, error) {
	coll := h.DB.MongoDB.Collection(collectionFeaturesCollectionName)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.HomeCollectionFeature
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func (h *HomeContentHandler) fetchTechCards(ctx context.Context) ([]models.TechShowcaseCard, error) {
	coll := h.DB.MongoDB.Collection(techCardsCollectionName)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.TechShowcaseCard
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func (h *HomeContentHandler) fetchTechHighlight(ctx context.Context) (*models.TechShowcaseHighlight, error) {
	coll := h.DB.MongoDB.Collection(techHighlightCollectionName)
	var highlight models.TechShowcaseHighlight
	err := coll.FindOne(ctx, bson.M{}).Decode(&highlight)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &highlight, nil
}

func (h *HomeContentHandler) clearHomeCache(ctx context.Context) {
	_ = h.DB.CacheDel(ctx, homeContentCacheKey)
}

func validateHeroSlide(slide *models.HeroSlide) error {
	if strings.TrimSpace(slide.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(slide.Subtitle) == "" {
		return errors.New("subtitle is required")
	}
	if strings.TrimSpace(slide.Description) == "" {
		return errors.New("description is required")
	}
	if strings.TrimSpace(slide.Image) == "" {
		return errors.New("image is required")
	}
	if slide.Features == nil {
		slide.Features = []string{}
	}
	return nil
}

func validateCategoryCard(card *models.HomeCategoryCard) error {
	if strings.TrimSpace(card.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(card.Subtitle) == "" {
		return errors.New("subtitle is required")
	}
	if strings.TrimSpace(card.Href) == "" {
		return errors.New("href is required")
	}
	if strings.TrimSpace(card.Image) == "" {
		return errors.New("image is required")
	}
	if strings.TrimSpace(card.BgGradient) == "" {
		return errors.New("bgGradient is required")
	}
	card.Href = strings.TrimSpace(card.Href)
	return nil
}

func validateCollectionFeature(feature *models.HomeCollectionFeature) error {
	if strings.TrimSpace(feature.Tagline) == "" {
		return errors.New("tagline is required")
	}
	if strings.TrimSpace(feature.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(feature.Description) == "" {
		return errors.New("description is required")
	}
	if strings.TrimSpace(feature.CtaLabel) == "" {
		return errors.New("ctaLabel is required")
	}
	if strings.TrimSpace(feature.CtaHref) == "" {
		return errors.New("ctaHref is required")
	}
	if strings.TrimSpace(feature.Image) == "" {
		return errors.New("image is required")
	}
	feature.CtaHref = strings.TrimSpace(feature.CtaHref)
	if strings.TrimSpace(feature.Layout) == "" {
		feature.Layout = "image-left"
	}
	return nil
}

func validateTechCard(card *models.TechShowcaseCard) error {
	if strings.TrimSpace(card.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(card.Subtitle) == "" {
		return errors.New("subtitle is required")
	}
	if strings.TrimSpace(card.Image) == "" && strings.TrimSpace(card.BackgroundImage) == "" {
		return errors.New("image or backgroundImage is required")
	}
	if strings.TrimSpace(card.Color) == "" {
		card.Color = "gray"
	}
	return nil
}

func validateHighlight(highlight *models.TechShowcaseHighlight) error {
	if strings.TrimSpace(highlight.Value) == "" {
		return errors.New("value is required")
	}
	if strings.TrimSpace(highlight.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(highlight.Subtitle) == "" {
		return errors.New("subtitle is required")
	}
	if strings.TrimSpace(highlight.AccentHex) == "" {
		highlight.AccentHex = "#f97316"
	}
	if strings.TrimSpace(highlight.Background) == "" {
		highlight.Background = "bg-rose-50"
	}
	return nil
}

func parseObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

func fiberError(c *fiber.Ctx, err error, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": message,
		"error":   err.Error(),
	})
}

func fiberBadRequest(c *fiber.Ctx, message string, err error) error {
	payload := fiber.Map{
		"success": false,
		"message": message,
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	return c.Status(fiber.StatusBadRequest).JSON(payload)
}

func fiberNotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
