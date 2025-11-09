# Product ID Fix for Home Content

## Problem

When clicking "Buy Now" on the home page hero slides or collection features, the wrong product details page was opening. This was because the system was trying to match products based on name, image, and price, which wasn't reliable.

## Solution

Added direct product ID linking to hero slides and collection features, so they can reference specific products directly.

## Backend Changes

### 1. Models Updated (`internal/models/home_content.go`)

Added `ProductID` and `Product` fields to:

- **HeroSlide**

  ```go
  ProductID   *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
  Product     *Product            `bson:"product,omitempty" json:"product,omitempty"`
  ```

- **HomeCollectionFeature**
  ```go
  ProductID    *primitive.ObjectID `bson:"productId,omitempty" json:"productId,omitempty"`
  Product      *Product            `bson:"product,omitempty" json:"product,omitempty"`
  ```

### 2. Handler Updates (`internal/handlers/home_content_handler.go`)

#### Updated `fetchHeroSlides()`:

- Now populates `Product` data when `ProductID` is set
- Fetches product from database and embeds it in the response

#### Updated `fetchCollectionFeatures()`:

- Now populates `Product` data when `ProductID` is set
- Fetches product from database and embeds it in the response

#### Updated `UpdateHeroSlide()`:

- Handles setting or clearing `productId` field
- Allows admins to link a specific product to a hero slide

#### Updated `UpdateCollectionFeature()`:

- Handles setting or clearing `productId` field
- Allows admins to link a specific product to a collection feature

## Frontend Changes

The frontend already had support for `productId` and `product` fields in the TypeScript types:

- `src/types/home-content.ts` - Types already defined
- `src/components/hero-content.tsx` - Already checks `productId` as Priority 1
- `src/components/collection.tsx` - Already checks `productId` as Priority 1

## How It Works Now

### Priority 1: Direct Product ID (NEW)

When an admin sets a `productId` on a hero slide or collection feature:

1. Backend fetches and embeds the full product data
2. Frontend receives `productId` and `product` in the API response
3. When user clicks "Buy Now", it navigates directly to: `/product_details?id={productId}`

### Priority 2: Matching Algorithm (Fallback)

If no `productId` is set, the frontend tries to match by:

- Name similarity
- Image filename matching
- Price matching
- Description matching

### Priority 3: Search by Name (Fallback)

Queries backend with the product name to find matches

### Priority 4: First Product (Fallback)

Shows the first available product

## How to Use (Admin)

### Setting Product ID in Database

You can link a product to a hero slide or collection feature by updating MongoDB directly or via API:

```javascript
// Example: Update hero slide with product ID
db.hero_slides.updateOne(
  { _id: ObjectId("slide_id_here") },
  {
    $set: {
      productId: ObjectId("product_id_here"),
    },
  }
);

// Example: Update collection feature with product ID
db.home_collection_features.updateOne(
  { _id: ObjectId("feature_id_here") },
  {
    $set: {
      productId: ObjectId("product_id_here"),
    },
  }
);
```

### Via API (PUT request)

```bash
# Update hero slide
curl -X PUT http://your-backend/admin/home-content/hero-slides/{slide_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "title": "MAK Chronograph",
    "subtitle": "Precision Timepiece",
    "productId": "product_object_id_here",
    ...other fields...
  }'

# Update collection feature
curl -X PUT http://your-backend/admin/home-content/collection-features/{feature_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "title": "Sport Collection",
    "tagline": "Built for Athletes",
    "productId": "product_object_id_here",
    ...other fields...
  }'
```

## Testing

1. **Without productId**: Click "Buy Now" on any hero slide or collection feature - should use the matching algorithm
2. **With productId**: Add a productId to a slide/feature in the database - click "Buy Now" should go directly to that product

## Benefits

1. ✅ **Accurate Navigation**: Users always see the correct product when clicking "Buy Now"
2. ✅ **Admin Control**: Admins can explicitly link products to promotional content
3. ✅ **Backward Compatible**: If no productId is set, falls back to the matching algorithm
4. ✅ **Performance**: Product data is pre-fetched and embedded in the API response
5. ✅ **SEO Friendly**: Direct product links improve crawlability

## API Response Example

```json
{
  "success": true,
  "message": "Home content retrieved successfully",
  "data": {
    "heroSlides": [
      {
        "id": "...",
        "title": "MAK Chronograph",
        "subtitle": "Precision Engineering",
        "productId": "507f1f77bcf86cd799439011",
        "product": {
          "id": "507f1f77bcf86cd799439011",
          "name": "MAK Chronograph Pro",
          "price": 45000,
          "images": [...],
          "stock": 10,
          ...
        },
        ...
      }
    ],
    "collections": [
      {
        "id": "...",
        "title": "Sport Collection",
        "productId": "507f1f77bcf86cd799439012",
        "product": {
          "id": "507f1f77bcf86cd799439012",
          "name": "MAK Sport Elite",
          "price": 35000,
          ...
        },
        ...
      }
    ]
  }
}
```

## Notes

- The `productId` field is optional - content can work without it
- When `productId` is set, `product` is automatically populated
- The product data is cached (5 minutes) along with home content
- Frontend prioritizes `productId` over all matching algorithms
