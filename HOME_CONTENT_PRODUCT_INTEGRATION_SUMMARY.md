# Home Content Product Integration - Implementation Summary

## Overview

This document outlines the implementation plan to integrate full product details into Hero Slides and Collection Features in the Home Content API. This will enable frontend to display product information, add to cart, and allow customers to purchase directly from home page sections.

---

## Current State Analysis

### Hero Slides (`HeroSlide` model)

**Location**: `internal/models/home_content.go`

**Current Fields**:

- `ID`, `Title`, `Subtitle`, `Price` (string), `Description`, `Image`
- `Features` ([]string), `Gradient`, `GlowColor`, `Position`
- `CreatedAt`, `UpdatedAt`

**Issues**:

- `Price` is a string (not structured)
- No product ID reference
- No product details (stock, category, etc.)
- Cannot be added to cart or purchased

### Collection Features (`HomeCollectionFeature` model)

**Location**: `internal/models/home_content.go`

**Current Fields**:

- `ID`, `Tagline`, `Title`, `Description`, `Availability`
- `CtaLabel`, `CtaHref`, `Image`, `ImageAlt`, `Layout`, `Position`
- `CreatedAt`, `UpdatedAt`

**Issues**:

- No product ID reference
- No pricing information
- No product details for cart functionality
- Only has a generic CTA href

---

## Proposed Solution

### 1. Update Data Models

#### A. Modify `HeroSlide` Model

**File**: `internal/models/home_content.go`

```go
type HeroSlide struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Title       string              `bson:"title" json:"title"`
	Subtitle    string              `bson:"subtitle" json:"subtitle"`
	Description string              `bson:"description" json:"description"`
	Image       string              `bson:"image" json:"image"`
	Features    []string            `bson:"features" json:"features"`
	Gradient    string              `bson:"gradient" json:"gradient"`
	GlowColor   string              `bson:"glowColor" json:"glowColor"`
	Position    int                 `bson:"position" json:"position"`

	// NEW: Product Integration Fields
	ProductID   *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
	Product     *Product            `bson:"-" json:"product,omitempty"` // Populated via lookup

	CreatedAt   time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time           `bson:"updatedAt" json:"updatedAt"`
}
```

**Changes**:

- Remove `Price` field (will come from Product)
- Add `ProductID` - optional reference to product
- Add `Product` - populated product details (not stored in DB, populated at runtime)

#### B. Modify `HomeCollectionFeature` Model

**File**: `internal/models/home_content.go`

```go
type HomeCollectionFeature struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Tagline      string              `bson:"tagline" json:"tagline"`
	Title        string              `bson:"title" json:"title"`
	Description  string              `bson:"description" json:"description"`
	Availability string              `bson:"availability" json:"availability"`
	CtaLabel     string              `bson:"ctaLabel" json:"ctaLabel"`
	CtaHref      string              `bson:"ctaHref" json:"ctaHref"`
	Image        string              `bson:"image" json:"image"`
	ImageAlt     string              `bson:"imageAlt" json:"imageAlt"`
	Layout       string              `bson:"layout" json:"layout"`
	Position     int                 `bson:"position" json:"position"`

	// NEW: Product Integration Fields
	ProductID    *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
	Product      *Product            `bson:"-" json:"product,omitempty"` // Populated via lookup

	CreatedAt    time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time           `bson:"updatedAt" json:"updatedAt"`
}
```

**Changes**:

- Add `ProductID` - optional reference to product
- Add `Product` - populated product details (not stored in DB, populated at runtime)
- Keep `CtaHref` for backward compatibility (can point to product detail page)

---

### 2. Update Handler Logic

#### A. Modify `fetchHeroSlides` Function

**File**: `internal/handlers/home_content_handler.go`

**Current Implementation** (line ~853):

```go
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
```

**New Implementation**:

```go
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

	// Populate product details for slides that have productId
	if err := h.populateHeroSlideProducts(ctx, slides); err != nil {
		// Log error but don't fail the request
		// Slides without products will just not have product data
	}

	return slides, nil
}

// New helper function
func (h *HomeContentHandler) populateHeroSlideProducts(ctx context.Context, slides []models.HeroSlide) error {
	productColl := h.DB.MongoDB.Collection("products")

	for i := range slides {
		if slides[i].ProductID != nil && !slides[i].ProductID.IsZero() {
			var product models.Product
			err := productColl.FindOne(ctx, bson.M{"_id": slides[i].ProductID}).Decode(&product)
			if err != nil {
				// Product not found or error - skip this one
				continue
			}
			slides[i].Product = &product
		}
	}

	return nil
}
```

#### B. Modify `fetchCollectionFeatures` Function

**File**: `internal/handlers/home_content_handler.go`

**Current Implementation** (line ~885):

```go
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
```

**New Implementation**:

```go
func (h *HomeContentHandler) fetchCollectionFeatures(ctx context.Context) ([]models.HomeCollectionFeature, error) {
	coll := h.DB.MongoDB.Collection(collectionFeaturesCollectionName)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var features []models.HomeCollectionFeature
	if err := cursor.All(ctx, &features); err != nil {
		return nil, err
	}

	// Populate product details for features that have productId
	if err := h.populateCollectionProducts(ctx, features); err != nil {
		// Log error but don't fail the request
	}

	return features, nil
}

// New helper function
func (h *HomeContentHandler) populateCollectionProducts(ctx context.Context, features []models.HomeCollectionFeature) error {
	productColl := h.DB.MongoDB.Collection("products")

	for i := range features {
		if features[i].ProductID != nil && !features[i].ProductID.IsZero() {
			var product models.Product
			err := productColl.FindOne(ctx, bson.M{"_id": features[i].ProductID}).Decode(&product)
			if err != nil {
				// Product not found or error - skip this one
				continue
			}
			features[i].Product = &product
		}
	}

	return nil
}
```

---

### 3. Update CRUD Operations

#### A. Update Validation Functions

**File**: `internal/handlers/home_content_handler.go`

**Update `validateHeroSlide`** (line ~962):

```go
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
	// ProductID is optional - no validation needed
	return nil
}
```

**Update `validateCollectionFeature`** (line ~1002):

```go
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
	// ProductID is optional - no validation needed
	return nil
}
```

#### B. Update Admin Create/Update Handlers

The existing handlers (`CreateHeroSlide`, `UpdateHeroSlide`, `CreateCollectionFeature`, `UpdateCollectionFeature`) will automatically work with the new `ProductID` field since they use `BodyParser` which will parse the JSON field if provided.

**No changes needed** - the handlers will accept `productId` in the request body automatically.

---

### 4. API Response Examples

#### A. Hero Slide with Product (GET /home-content)

**Response**:

```json
{
  "success": true,
  "message": "Home content retrieved successfully",
  "data": {
    "heroSlides": [
      {
        "id": "674567890abcdef123456789",
        "title": "Luminous Chronograph",
        "subtitle": "Precision Meets Elegance",
        "description": "Experience the perfect blend of sophisticated design and cutting-edge technology",
        "image": "https://s3.amazonaws.com/hero-images/luminous-chronograph.jpg",
        "features": [
          "Swiss Movement",
          "Water Resistant 100m",
          "Sapphire Crystal"
        ],
        "gradient": "from-blue-900 to-purple-900",
        "glowColor": "blue",
        "position": 1,
        "productId": "673456789abcdef012345678",
        "product": {
          "id": "673456789abcdef012345678",
          "name": "Luminous Chronograph Elite",
          "brand": "MAK Watches",
          "description": "Premium chronograph with Swiss movement",
          "price": 24999.0,
          "category": "Watches",
          "mainCategory": "Luxury",
          "subcategory": "Chronograph",
          "imageUrl": "https://s3.amazonaws.com/products/luminous-elite.jpg",
          "images": [
            "https://s3.amazonaws.com/products/luminous-elite-1.jpg",
            "https://s3.amazonaws.com/products/luminous-elite-2.jpg"
          ],
          "stock": 15,
          "gender": "Men",
          "dialColor": "Blue",
          "dialShape": "Round",
          "dialType": "Chronograph",
          "strapColor": "Black",
          "strapMaterial": "Leather",
          "discountPercentage": 10,
          "createdAt": "2024-11-01T10:00:00Z",
          "updatedAt": "2024-11-08T15:30:00Z"
        },
        "createdAt": "2024-11-05T08:00:00Z",
        "updatedAt": "2024-11-09T09:00:00Z"
      }
    ],
    "collections": [
      {
        "id": "674567890abcdef123456790",
        "tagline": "Limited Edition",
        "title": "Ocean Series",
        "description": "Dive into luxury with our water-resistant collection",
        "availability": "Only 50 pieces worldwide",
        "ctaLabel": "Explore Collection",
        "ctaHref": "/products/673456789abcdef012345679",
        "image": "https://s3.amazonaws.com/collections/ocean-series.jpg",
        "imageAlt": "Ocean Series Watch",
        "layout": "image-left",
        "position": 1,
        "productId": "673456789abcdef012345679",
        "product": {
          "id": "673456789abcdef012345679",
          "name": "Ocean Diver Pro",
          "brand": "MAK Watches",
          "description": "Professional diving watch with 300m water resistance",
          "price": 34999.0,
          "category": "Watches",
          "mainCategory": "Sports",
          "subcategory": "Diving",
          "imageUrl": "https://s3.amazonaws.com/products/ocean-diver.jpg",
          "images": [
            "https://s3.amazonaws.com/products/ocean-diver-1.jpg",
            "https://s3.amazonaws.com/products/ocean-diver-2.jpg"
          ],
          "stock": 50,
          "gender": "Unisex",
          "dialColor": "Black",
          "dialShape": "Round",
          "dialType": "Analog",
          "strapColor": "Blue",
          "strapMaterial": "Rubber",
          "discountPercentage": 15,
          "createdAt": "2024-11-01T10:00:00Z",
          "updatedAt": "2024-11-08T15:30:00Z"
        },
        "createdAt": "2024-11-05T08:00:00Z",
        "updatedAt": "2024-11-09T09:00:00Z"
      }
    ]
  }
}
```

#### B. Admin Create Hero Slide with Product (POST /admin/home-content/hero-slides)

**Request Body**:

```json
{
  "title": "Titanium Elite",
  "subtitle": "Lightweight Performance",
  "description": "Experience unmatched comfort with aerospace-grade titanium",
  "image": "https://s3.amazonaws.com/hero-images/titanium-elite.jpg",
  "features": ["Titanium Case", "Anti-Magnetic", "70-hour Power Reserve"],
  "gradient": "from-gray-800 to-slate-900",
  "glowColor": "silver",
  "position": 2,
  "productId": "673456789abcdef012345680"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Hero slide created successfully",
  "data": {
    "id": "674567890abcdef123456791",
    "title": "Titanium Elite",
    "subtitle": "Lightweight Performance",
    "description": "Experience unmatched comfort with aerospace-grade titanium",
    "image": "https://s3.amazonaws.com/hero-images/titanium-elite.jpg",
    "features": ["Titanium Case", "Anti-Magnetic", "70-hour Power Reserve"],
    "gradient": "from-gray-800 to-slate-900",
    "glowColor": "silver",
    "position": 2,
    "productId": "673456789abcdef012345680",
    "createdAt": "2024-11-09T10:30:00Z",
    "updatedAt": "2024-11-09T10:30:00Z"
  }
}
```

---

## Implementation Steps

### Step 1: Update Models

1. Open `internal/models/home_content.go`
2. Modify `HeroSlide` struct:
   - Remove `Price string` field
   - Add `ProductID *primitive.ObjectID` field
   - Add `Product *Product` field
3. Modify `HomeCollectionFeature` struct:
   - Add `ProductID *primitive.ObjectID` field
   - Add `Product *Product` field

### Step 2: Update Handler Fetch Functions

1. Open `internal/handlers/home_content_handler.go`
2. Add `populateHeroSlideProducts` helper function
3. Modify `fetchHeroSlides` to call the populate function
4. Add `populateCollectionProducts` helper function
5. Modify `fetchCollectionFeatures` to call the populate function

### Step 3: Test Existing CRUD Operations

The existing admin CRUD operations should work automatically since:

- `BodyParser` will parse `productId` field from JSON
- MongoDB will store the new field
- Validation functions don't require `productId` (it's optional)

### Step 4: Update Database (Migration)

**Optional**: If there are existing hero slides with a `price` field, you may want to:

1. Create a migration script to remove the `price` field from existing documents
2. Or leave it (it won't be returned in the API response)

**MongoDB Migration Example**:

```javascript
// Remove price field from hero_slides
db.hero_slides.updateMany({}, { $unset: { price: "" } });

// Optionally add productId to existing slides if you know the product IDs
db.hero_slides.updateOne(
  { title: "Luminous Chronograph" },
  { $set: { productId: ObjectId("673456789abcdef012345678") } }
);
```

### Step 5: Frontend Integration

Frontend should:

1. Check if `product` field exists in hero slide/collection
2. If exists, display product details (name, price, stock)
3. Use product ID to enable "Add to Cart" functionality
4. Link to product detail page using product ID
5. Display final price considering any active discounts (use `product.discountPercentage` or `product.discountAmount`)

**Frontend Example (React/Angular)**:

```typescript
// Hero Slide Component
interface HeroSlide {
  id: string;
  title: string;
  subtitle: string;
  description: string;
  image: string;
  features: string[];
  gradient: string;
  glowColor: string;
  productId?: string;
  product?: Product;
}

// In template
<div *ngIf="slide.product" class="product-info">
  <h3>{{ slide.product.name }}</h3>
  <p class="price">
    <span *ngIf="slide.product.discountPercentage" class="original">
      ₹{{ slide.product.price }}
    </span>
    ₹{{ calculateFinalPrice(slide.product) }}
  </p>
  <p class="stock">{{ slide.product.stock }} in stock</p>
  <button (click)="addToCart(slide.product.id)">Add to Cart</button>
  <a [routerLink]="['/products', slide.product.id]">View Details</a>
</div>

// Helper function
calculateFinalPrice(product: Product): number {
  if (!product.discountPercentage) return product.price;
  return product.price - (product.price * product.discountPercentage / 100);
}
```

---

## API Endpoints Summary

### Public Endpoints

| Method | Endpoint        | Description                                  |
| ------ | --------------- | -------------------------------------------- |
| GET    | `/home-content` | Get all home content with populated products |

### Admin Endpoints (Requires Authentication + Admin Role)

| Method | Endpoint                              | Description                                   |
| ------ | ------------------------------------- | --------------------------------------------- |
| GET    | `/admin/home-content/hero-slides`     | List all hero slides                          |
| POST   | `/admin/home-content/hero-slides`     | Create hero slide (with optional `productId`) |
| PUT    | `/admin/home-content/hero-slides/:id` | Update hero slide (with optional `productId`) |
| DELETE | `/admin/home-content/hero-slides/:id` | Delete hero slide                             |
| GET    | `/admin/home-content/collections`     | List all collection features                  |
| POST   | `/admin/home-content/collections`     | Create collection (with optional `productId`) |
| PUT    | `/admin/home-content/collections/:id` | Update collection (with optional `productId`) |
| DELETE | `/admin/home-content/collections/:id` | Delete collection                             |

---

## Benefits

### 1. **E-commerce Functionality**

- Users can add products to cart directly from hero slides
- Users can purchase featured collection products immediately
- Seamless shopping experience without leaving home page

### 2. **Rich Product Information**

- Display real-time pricing with active discounts
- Show stock availability
- Display product images and details
- Support all product attributes (dial color, strap material, etc.)

### 3. **Centralized Product Management**

- Single source of truth for product data
- Price updates automatically reflect in home content
- Inventory management applies to hero/collection products
- Discount campaigns automatically apply

### 4. **Backward Compatibility**

- `productId` is optional - existing content without products still works
- Can mix product-linked and non-product content
- Frontend can gracefully handle missing product data

### 5. **Performance**

- Products are fetched once per request
- Cached with home content (5 minutes)
- Efficient database queries

---

## Testing Checklist

### API Testing

- [ ] GET `/home-content` returns hero slides with populated products
- [ ] GET `/home-content` returns collections with populated products
- [ ] GET `/home-content` works when productId is null
- [ ] GET `/home-content` works when productId references non-existent product
- [ ] POST `/admin/home-content/hero-slides` accepts productId field
- [ ] POST `/admin/home-content/hero-slides` works without productId
- [ ] PUT `/admin/home-content/hero-slides/:id` can update productId
- [ ] PUT `/admin/home-content/hero-slides/:id` can remove productId (set to null)
- [ ] POST `/admin/home-content/collections` accepts productId field
- [ ] PUT `/admin/home-content/collections/:id` can update productId
- [ ] Cache invalidation works after creating/updating content

### Frontend Testing

- [ ] Hero slide displays product name and price
- [ ] "Add to Cart" button works for hero slide products
- [ ] Product link navigates to correct product detail page
- [ ] Discount pricing displays correctly
- [ ] Stock availability shows correctly
- [ ] Collection feature displays product details
- [ ] "Add to Cart" works for collection products
- [ ] Graceful handling when product data is missing

### Database Testing

- [ ] ProductID field stores correctly in MongoDB
- [ ] ProductID can be null/omitted
- [ ] Product lookup works via ObjectID reference
- [ ] Migration removes old price field (if applicable)

---

## Performance Considerations

1. **Caching**: Home content is cached for 5 minutes, including populated products
2. **Lazy Loading**: Products are only fetched if `productId` is present
3. **Error Handling**: Missing products don't break the entire home content response
4. **N+1 Query Prevention**: Could be optimized later with bulk product fetch if needed

### Future Optimization (Optional)

If there are many hero slides/collections with products, consider batch fetching:

```go
// Collect all product IDs
var productIDs []primitive.ObjectID
for _, slide := range slides {
    if slide.ProductID != nil && !slide.ProductID.IsZero() {
        productIDs = append(productIDs, *slide.ProductID)
    }
}

// Batch fetch all products
products := make(map[primitive.ObjectID]*models.Product)
if len(productIDs) > 0 {
    cursor, _ := productColl.Find(ctx, bson.M{"_id": bson.M{"$in": productIDs}})
    var productList []models.Product
    cursor.All(ctx, &productList)
    for i := range productList {
        products[productList[i].ID] = &productList[i]
    }
}

// Assign products to slides
for i := range slides {
    if slides[i].ProductID != nil {
        if prod, exists := products[*slides[i].ProductID]; exists {
            slides[i].Product = prod
        }
    }
}
```

---

## Summary

This implementation adds full e-commerce functionality to Home Content sections by:

1. ✅ Linking hero slides and collections to actual products
2. ✅ Populating complete product details (price, stock, images, etc.)
3. ✅ Enabling "Add to Cart" and purchase functionality
4. ✅ Maintaining backward compatibility
5. ✅ Keeping the API simple and performant

The changes are minimal, non-breaking, and leverage existing product infrastructure!
