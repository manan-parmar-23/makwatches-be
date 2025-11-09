# Home Content Cart Support Documentation

## Overview

This document explains how home content products (hero slides and collection features) are now fully integrated with the cart and checkout system, allowing users to add home page products directly to cart and complete purchases.

## Changes Made

### 1. Backend Model Updates

#### `internal/models/home_content.go`

Added `Price` field to `HomeCollectionFeature` model:

```go
type HomeCollectionFeature struct {
    // ... existing fields ...
    Price        string  `bson:"price,omitempty" json:"price,omitempty"`
    // ... rest of fields ...
}
```

### 2. Cart Handler Updates

#### `internal/handlers/cart_handler.go`

**AddToCart Function:**

- Now checks hero_slides and home_collection_features collections first (by `productId` field)
- Falls back to regular products collection (by `_id` field)
- Home content products have unlimited stock (no stock validation)

**GetCart Function:**

- Retrieves product details from all three sources:
  1. Regular products collection
  2. Hero slides collection
  3. Collection features collection
- Parses price from string format (e.g., "₹3800") to float for home content products
- Converts home content items to Product format for consistent cart display

### 3. Home Content Handler Updates

#### `internal/handlers/home_content_handler.go`

**GetHomeContentByProductID Function:**

- Added price field to collection feature response
- Ensures consistent product data structure for both hero slides and collection features

## Data Flow

### Adding Home Content Product to Cart

```
User clicks "Buy Now" on Hero Slide
    ↓
Frontend sends POST /cart with productId (e.g., 6910752af5596c695638e7f5)
    ↓
Backend checks hero_slides collection WHERE productId = 6910752af5596c695638e7f5
    ↓
Found? → Add to cart with productId
Not Found? → Check home_collection_features
    ↓
Not Found? → Check products collection WHERE _id = 6910752af5596c695638e7f5
    ↓
Not Found? → Return 404 "Product not found"
```

### Retrieving Cart with Home Content Products

```
User views cart
    ↓
Frontend sends GET /cart/:userID
    ↓
Backend fetches all cart items for user
    ↓
For each cart item:
    1. Try products collection (by _id)
    2. If not found, try hero_slides (by productId)
    3. If not found, try home_collection_features (by productId)
    ↓
Convert all items to Product format
    ↓
Return cart with total price
```

## Product ID Strategy

### Home Content Products

- Each hero slide and collection feature has a unique `productId` (ObjectID)
- This `productId` is auto-generated on create/update
- Cart stores items using this `productId`
- Product details page fetches using this `productId`

### Regular Shop Products

- Shop products use `_id` as identifier
- Cart can also store regular products by `_id`
- System supports both types simultaneously

## Price Handling

### Home Content Price Format

```go
// Hero Slide / Collection Feature
Price: "₹3800"  // String format

// Parsing in GetCart
fmt.Sscanf(heroSlide.Price, "₹%f", &priceFloat)
// or
fmt.Sscanf(heroSlide.Price, "%f", &priceFloat)

// Result
priceFloat: 3800.0  // Float format for cart calculation
```

### Regular Product Price Format

```go
// Product
Price: 3800.0  // Float format (already compatible)
```

## Testing Instructions

### 1. Test Adding Hero Slide to Cart

```bash
# Navigate to home page
# Click "Buy Now" on any hero slide
# Should see success message

# Verify in cart API
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://127.0.0.1:8080/cart/YOUR_USER_ID
```

### 2. Test Adding Collection Feature to Cart

```bash
# Navigate to home page
# Scroll to collection features
# Click CTA button (if linked to product details)
# Click "Add to Cart"
# Should see success message
```

### 3. Test Checkout Flow

```bash
# Add home content product to cart
# Proceed to checkout
# Complete payment
# Order should include home content product
```

### 4. Test Mixed Cart (Regular + Home Content)

```bash
# Add regular shop product to cart
# Add hero slide product to cart
# View cart
# Both should display correctly with proper prices
# Proceed to checkout
# Order should include both types
```

## Database Examples

### Hero Slide with ProductID

```json
{
  "_id": ObjectId("690e4dfadbacdf15a21cdec"),
  "title": "MAK Watches",
  "subtitle": "Fastrack Vyb",
  "price": "₹3800",
  "description": "Premium smartwatch",
  "image": "https://example.com/image.jpg",
  "productId": ObjectId("6910752af5596c695638e7f5"),
  "features": ["Water resistant", "GPS tracking"],
  "createdAt": ISODate("2025-01-15T10:00:00Z"),
  "updatedAt": ISODate("2025-01-15T10:00:00Z")
}
```

### Cart Item Referencing Home Content

```json
{
  "_id": ObjectId("..."),
  "user_id": ObjectId("68f4f625700b52a9a9b2cc67"),
  "product_id": ObjectId("6910752af5596c695638e7f5"),
  "quantity": 1,
  "created_at": ISODate("2025-01-15T10:30:00Z"),
  "updated_at": ISODate("2025-01-15T10:30:00Z")
}
```

### GetCart Response with Home Content Product

```json
{
  "success": true,
  "message": "Cart retrieved successfully",
  "data": {
    "items": [
      {
        "_id": "...",
        "user_id": "68f4f625700b52a9a9b2cc67",
        "product_id": "6910752af5596c695638e7f5",
        "quantity": 1,
        "product": {
          "id": "6910752af5596c695638e7f5",
          "name": "MAK Watches",
          "description": "Premium smartwatch",
          "price": 3800,
          "imageUrl": "https://example.com/image.jpg",
          "images": ["https://example.com/image.jpg"],
          "stock": 999
        }
      }
    ],
    "total": 3800
  }
}
```

## API Endpoints

### Cart Operations

- `POST /cart` - Add product to cart (works with home content productId)
- `GET /cart/:userID` - Get cart (includes home content products)
- `DELETE /cart/:userID/:productID` - Remove from cart (works with home content productId)

### Home Content Product Retrieval

- `GET /home-content/product/:productId` - Get hero slide or collection by productId

### Regular Product Retrieval

- `GET /catalog/products/:id` - Get shop product by id

## Benefits

1. **Unified Shopping Experience**: Users can purchase from home page without navigating to shop
2. **Flexible Product Management**: Admin can create promotional products directly in home content
3. **No Stock Limitations**: Home content products have unlimited stock by default
4. **Backward Compatible**: Regular shop products continue to work as before
5. **Single Checkout Flow**: Both product types use the same cart and checkout process

## Known Limitations

1. Home content products don't have detailed specifications (dial color, strap material, etc.)
2. Price is stored as string in home content (automatically converted in cart)
3. Home content products always show as "in stock" (999 units)
4. No discount/promotion support for home content products yet

## Future Enhancements

1. Add discount support for home content products
2. Add detailed specifications to home content products
3. Add stock management for home content products
4. Add variant support (size, color) for home content products
5. Add product reviews for home content products

## Troubleshooting

### Issue: Cart returns 404 when adding home content product

**Solution:** Ensure the hero slide or collection has a `productId` field. Update the item through admin panel to auto-generate productId.

### Issue: Cart displays wrong price for home content product

**Solution:** Check the price format in the database. It should be a string with or without currency symbol (e.g., "₹3800" or "3800").

### Issue: Product details page shows fallback data

**Solution:** Verify that the `/home-content/product/:productId` endpoint returns the correct data. Check backend logs for errors.

### Issue: Can't checkout with home content product

**Solution:** Ensure the order handler supports mixed cart items. Home content products should be treated like regular products during checkout.

---

**Last Updated:** January 15, 2025
**Version:** 1.0
