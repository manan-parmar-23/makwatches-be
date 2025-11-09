# Product ID Auto-Generation Fix

## Overview

Fixed the issue where clicking "Buy Now" on hero slides and collection features opens the wrong product by auto-generating unique `productId` for each item.

## Solution Approach

Instead of trying to match hero slides/collections to existing shop products, each hero slide and collection feature now gets its own unique `productId` that serves as a direct identifier. The product details page can fetch content from either:

1. Regular shop products (`/catalog/products/:id`)
2. Home content (hero slides/collections) (`/home-content/product/:productId`)

## Backend Changes

### 1. Auto-Generate ProductID on Creation

**File:** `internal/handlers/home_content_handler.go`

#### CreateHeroSlide

```go
// Generate productId if not provided
if payload.ProductID == nil || payload.ProductID.IsZero() {
    newProductID := primitive.NewObjectID()
    payload.ProductID = &newProductID
}
```

#### CreateCollectionFeature

```go
// Generate productId if not provided
if payload.ProductID == nil || payload.ProductID.IsZero() {
    newProductID := primitive.NewObjectID()
    payload.ProductID = &newProductID
}
```

### 2. Auto-Generate ProductID on Update (if missing)

**File:** `internal/handlers/home_content_handler.go`

#### UpdateHeroSlide

```go
// Fetch existing slide first
var existing models.HeroSlide
err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existing)

// Generate productId if it doesn't exist yet
if existing.ProductID == nil || existing.ProductID.IsZero() {
    newProductID := primitive.NewObjectID()
    update["productId"] = newProductID
}
```

#### UpdateCollectionFeature

```go
// Fetch existing feature first
var existing models.HomeCollectionFeature
err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existing)

// Generate productId if it doesn't exist yet
if existing.ProductID == nil || existing.ProductID.IsZero() {
    newProductID := primitive.NewObjectID()
    update["productId"] = newProductID
}
```

### 3. New Endpoint: Get Home Content by ProductID

**File:** `internal/handlers/home_content_handler.go`

**Endpoint:** `GET /home-content/product/:productId`

This endpoint searches both hero slides and collection features for a matching `productId` and returns the content in a product-compatible format.

```go
func (h *HomeContentHandler) GetHomeContentByProductID(c *fiber.Ctx) error {
    // Search hero slides first
    var heroSlide models.HeroSlide
    err = heroCollection.FindOne(ctx, bson.M{"productId": objID}).Decode(&heroSlide)
    if err == nil {
        // Return hero slide as product data
        return product-like response
    }

    // Search collection features
    var collectionFeature models.HomeCollectionFeature
    err = collectionCollection.FindOne(ctx, bson.M{"productId": objID}).Decode(&collectionFeature)
    if err == nil {
        // Return collection as product data
        return product-like response
    }

    return "Product not found"
}
```

**Response Format:**

```json
{
  "success": true,
  "message": "Product details retrieved from hero content",
  "data": {
    "id": "generated_product_id",
    "name": "MAK Watches",
    "description": "Fastrack Vyb Limitless...",
    "price": "₹3800",
    "images": ["https://..."],
    "subtitle": "Fastrack Vyb",
    "features": ["12 Month Warranty", "..."],
    "source": "hero_slide",
    "sourceId": "original_slide_id"
  }
}
```

### 4. Route Registration

**File:** `internal/handlers/handlers.go`

```go
app.Get("/home-content/product/:productId", homeContentHandler.GetHomeContentByProductID)
```

## Frontend Changes

### Product Details Page

**File:** `src/app/product_details/page.tsx`

Updated to try fetching from both sources:

```typescript
// First try regular products
try {
  response = await fetchPublicProductById(id);
} catch (productError) {
  // If not found, try home content
  try {
    const homeContentUrl = `${API_URL}/home-content/product/${id}`;
    response = await fetch(homeContentUrl);
    // Convert to expected format
  } catch {
    throw productError;
  }
}
```

Also handles price parsing for both number and string formats:

```typescript
// Parse price if it's a string (from home content)
let priceValue = FALLBACK.price;
if (typeof p.price === "number") {
  priceValue = p.price;
} else if (p.price && typeof p.price === "string") {
  const parsed = parseFloat((p.price as string).replace(/[^0-9.]/g, ""));
  if (!isNaN(parsed)) {
    priceValue = parsed;
  }
}
```

## Data Flow

### On Creation/Update:

```
1. Admin creates/updates hero slide or collection feature
   ↓
2. Backend checks if productId exists
   ↓
3. If not, generates new ObjectID
   ↓
4. Saves to database with productId
   ↓
5. Returns data including productId to frontend
```

### On "Buy Now" Click:

```
1. User clicks "Buy Now" on hero slide
   ↓
2. Frontend uses slide.productId
   ↓
3. Navigates to: /product_details?id={productId}
   ↓
4. Product details page tries /catalog/products/{id}
   ↓
5. If 404, tries /home-content/product/{id}
   ↓
6. Displays correct product details
```

## Example Data

### Before (Existing Slide):

```javascript
{
  _id: ObjectId("690e4dfadbacdf15a21cdec"),
  title: "MAK Watches",
  subtitle: "Fastrack Vyb",
  price: "₹3800",
  // ❌ No productId
}
```

### After Update:

```javascript
{
  _id: ObjectId("690e4dfadbacdf15a21cdec"),
  title: "MAK Watches",
  subtitle: "Fastrack Vyb",
  price: "₹3800",
  productId: ObjectId("673fa1b2c8d4e5f6a7b8c9d0"), // ✅ Auto-generated
  updatedAt: ISODate("2025-11-09...")
}
```

### API Response:

```json
{
  "id": "690e4dfadbacdf15a21cdec",
  "title": "MAK Watches",
  "subtitle": "Fastrack Vyb",
  "productId": "673fa1b2c8d4e5f6a7b8c9d0", // ✅ Included
  ...
}
```

## Benefits

✅ **Simple**: No complex matching algorithms needed
✅ **Reliable**: Each item has its own unique identifier
✅ **Consistent**: Product details page works for both shop products and home content
✅ **Automatic**: ProductID generated on create/update without manual intervention
✅ **Backward Compatible**: Existing slides get productId when updated
✅ **Data Integrity**: Each hero slide/collection is its own product entity

## Testing

### 1. Test Existing Slide

```bash
# Update an existing slide (will generate productId)
curl -X PUT http://localhost:8080/admin/home-content/hero-slides/{slide_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "MAK Watches",
    "subtitle": "Fastrack Vyb",
    "price": "₹3800",
    ...
  }'

# Response will include generated productId
```

### 2. Test Product Details

```bash
# Use the productId from the slide
curl http://localhost:8080/home-content/product/{productId}

# Should return product-like data from the slide
```

### 3. Test Frontend

1. Go to home page
2. Click "Buy Now" on any hero slide
3. Should open product details with correct information
4. URL: `/product_details?id={generated_productId}`

## Files Modified

- `internal/handlers/home_content_handler.go` - Auto-generate productId, new endpoint
- `internal/handlers/handlers.go` - Route registration
- `internal/models/home_content.go` - Already has ProductID field
- `src/app/product_details/page.tsx` - Fetch from home content as fallback

## Summary

The solution automatically generates a unique `productId` for every hero slide and collection feature when created or updated. The product details page can fetch from either shop products or home content, ensuring the correct product always displays when clicking "Buy Now".

No manual database updates needed - just update any slide and it will get a productId automatically!
