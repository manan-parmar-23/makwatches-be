package handlers

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/models"
)

// CategoryHandler handles category management
// Routes are mounted under /admin/categories with admin middleware.
type CategoryHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewCategoryHandler creates a new handler instance
func NewCategoryHandler(db *database.DBClient, cfg *config.Config) *CategoryHandler {
	return &CategoryHandler{DB: db, Config: cfg}
}

// CreateCategory creates a main category (Men or Women) with optional subcategories
// @example Request:
// POST /admin/categories
//
//	{
//	  "name": "Men",
//	  "subcategories": ["Shirts", "Jeans"]
//	}
//
// @example Response (201):
//
//	{
//	  "success": true,
//	  "message": "Category created successfully",
//	  "data": {"id": "...","name": "Men","subcategories": [{"id": "...","name": "Shirts"}],"createdAt": "...","updatedAt": "..."}
//	}
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	// Parse payload allowing subcategories to be either []string or []SubcategoryInput
	var raw struct {
		Name          string          `json:"name"`
		Subcategories json.RawMessage `json:"subcategories"`
	}
	if err := c.BodyParser(&raw); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid request body", "error": err.Error()})
	}

	if raw.Name != "Men" && raw.Name != "Women" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Category name must be 'Men' or 'Women'"})
	}

	collection := h.DB.Collections().Categories

	// Ensure category uniqueness (Men/Women only once)
	count, err := collection.CountDocuments(ctx, bson.M{"name": raw.Name})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Database error", "error": err.Error()})
	}
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Category already exists"})
	}

	now := time.Now()
	subcats := make([]models.Subcategory, 0)
	if len(raw.Subcategories) > 0 && string(raw.Subcategories) != "null" {
		// Try []string first
		var names []string
		if err := json.Unmarshal(raw.Subcategories, &names); err == nil {
			for _, s := range names {
				if s == "" {
					continue
				}
				subcats = append(subcats, models.Subcategory{ID: primitive.NewObjectID(), Name: s})
			}
		} else {
			// Try []SubcategoryInput
			var inputs []models.SubcategoryInput
			if err2 := json.Unmarshal(raw.Subcategories, &inputs); err2 == nil {
				for _, in := range inputs {
					if in.Name == "" {
						continue
					}
					subcats = append(subcats, models.Subcategory{ID: primitive.NewObjectID(), Name: in.Name, ImageURL: in.ImageURL})
				}
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid subcategories format"})
			}
		}
	}

	cat := models.Category{
		ID:            primitive.NewObjectID(),
		Name:          raw.Name,
		Subcategories: subcats,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := collection.InsertOne(ctx, cat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create category", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "message": "Category created successfully", "data": cat})
}

// AddSubcategory adds a subcategory to an existing category
// @example Request:
// POST /admin/categories/:id/subcategories
// {"name": "Shoes"}
// @example Response (200):
// {"success": true,"message": "Subcategory added successfully","data": {"id": "...","name": "Men","subcategories": [{"id": "...","name": "Shoes"}],"createdAt": "...","updatedAt": "..."}}
func (h *CategoryHandler) AddSubcategory(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid category id"})
	}

	var req models.AddSubcategoryRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid subcategory"})
	}

	collection := h.DB.Collections().Categories

	subcat := models.Subcategory{ID: primitive.NewObjectID(), Name: req.Name, ImageURL: req.ImageURL}
	update := bson.M{
		"$push": bson.M{"subcategories": subcat},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts)
	var updated models.Category
	if err := res.Decode(&updated); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Subcategory added successfully", "data": updated})
}

// UpdateCategoryName updates a main category name (switch Men/Women)
// PATCH /admin/categories/:id
// {"name": "Women"}
func (h *CategoryHandler) UpdateCategoryName(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid category id"})
	}

	var req models.UpdateNameRequest
	if err := c.BodyParser(&req); err != nil || (req.Name != "Men" && req.Name != "Women") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Name must be 'Men' or 'Women'"})
	}

	collection := h.DB.Collections().Categories

	update := bson.M{"$set": bson.M{"name": req.Name, "updated_at": time.Now()}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts)
	var updated models.Category
	if err := res.Decode(&updated); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Category updated successfully", "data": updated})
}

// UpdateSubcategoryName updates a subcategory name
// PATCH /admin/categories/:categoryId/subcategories/:subId
// {"name": "Sneakers"}
func (h *CategoryHandler) UpdateSubcategoryName(c *fiber.Ctx) error {
	ctx := c.Context()
	categoryID := c.Params("categoryId")
	subID := c.Params("subId")

	catObj, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid category id"})
	}
	subObj, err := primitive.ObjectIDFromHex(subID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid subcategory id"})
	}

	// Accept payloads to update name and/or imageUrl
	var req models.UpdateSubcategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid payload"})
	}
	if (req.Name == nil || *req.Name == "") && req.ImageURL == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nothing to update"})
	}

	collection := h.DB.Collections().Categories

	filter := bson.M{"_id": catObj, "subcategories._id": subObj}
	set := bson.M{"updated_at": time.Now()}
	if req.Name != nil && *req.Name != "" {
		set["subcategories.$.name"] = *req.Name
	}
	if req.ImageURL != nil {
		// Allow clearing image when empty string provided
		if *req.ImageURL == "" {
			set["subcategories.$.image_url"] = nil
		} else {
			set["subcategories.$.image_url"] = *req.ImageURL
		}
	}
	update := bson.M{"$set": set}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, filter, update, opts)
	var updated models.Category
	if err := res.Decode(&updated); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category or subcategory not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Subcategory updated successfully", "data": updated})
}

// DeleteCategory deletes a category entirely
// DELETE /admin/categories/:id
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid category id"})
	}

	collection := h.DB.Collections().Categories
	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to delete category", "error": err.Error()})
	}
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Category deleted successfully"})
}

// DeleteSubcategory removes a subcategory from a category
// DELETE /admin/categories/:categoryId/subcategories/:subId
func (h *CategoryHandler) DeleteSubcategory(c *fiber.Ctx) error {
	ctx := c.Context()
	categoryID := c.Params("categoryId")
	subID := c.Params("subId")
	catObj, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid category id"})
	}
	subObj, err := primitive.ObjectIDFromHex(subID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid subcategory id"})
	}

	collection := h.DB.Collections().Categories
	update := bson.M{"$pull": bson.M{"subcategories": bson.M{"_id": subObj}}, "$set": bson.M{"updated_at": time.Now()}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, bson.M{"_id": catObj}, update, opts)
	var updated models.Category
	if err := res.Decode(&updated); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category or subcategory not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Subcategory deleted successfully", "data": updated})
}

// GetCategories fetches all categories
// GET /admin/categories
func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	ctx := c.Context()
	collection := h.DB.Collections().Categories

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch categories", "error": err.Error()})
	}
	defer cursor.Close(ctx)

	var cats []models.Category
	if err := cursor.All(ctx, &cats); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to decode categories", "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Categories retrieved successfully", "data": cats})
}

// GetPublicCategories provides a public (non-admin, no-auth) list of categories.
// GET /categories?name=Men|Women (optional filter)
// Always returns success=true with an array (possibly empty) for easier client handling.
func (h *CategoryHandler) GetPublicCategories(c *fiber.Ctx) error {
	ctx := c.Context()
	collection := h.DB.Collections().Categories

	filter := bson.M{}
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch categories", "error": err.Error()})
	}
	defer cursor.Close(ctx)

	var cats []models.Category
	if err := cursor.All(ctx, &cats); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to decode categories", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Categories retrieved successfully", "data": cats})
}

// GetPublicSubcategories returns only the subcategories for a given main category by name.
// GET /categories/:name/subcategories
// Returns 200 with empty list if category not found (avoids leaking existence semantics) unless strict is requested via ?strict=1.
func (h *CategoryHandler) GetPublicSubcategories(c *fiber.Ctx) error {
	ctx := c.Context()
	name := c.Params("name")
	collection := h.DB.Collections().Categories

	var cat models.Category
	err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&cat)
	if err != nil {
		if err == fiber.ErrNotFound || err.Error() == "mongo: no documents in result" {
			if c.Query("strict") == "1" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Category not found"})
			}
			return c.JSON(fiber.Map{"success": true, "message": "Subcategories retrieved successfully", "data": []models.Subcategory{}})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch category", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Subcategories retrieved successfully", "data": cat.Subcategories})
}
