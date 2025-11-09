// Helper script to link products to hero slides and collection features
// Run in MongoDB shell or MongoDB Compass

// ===========================================
// EXAMPLE 1: Link a product to a hero slide
// ===========================================

// First, find the product you want to link
db.products.findOne({ name: /chronograph/i })
// Copy the _id from the result

// Then update the hero slide
db.hero_slides.updateOne(
  { title: "MAK Chronograph" }, // Find by title
  { 
    $set: { 
      productId: ObjectId("PASTE_PRODUCT_ID_HERE") 
    } 
  }
)

// Verify the update
db.hero_slides.findOne({ title: "MAK Chronograph" })

// ===========================================
// EXAMPLE 2: Link a product to a collection feature
// ===========================================

// Find the product
db.products.findOne({ name: /sport/i })
// Copy the _id

// Update the collection feature
db.home_collection_features.updateOne(
  { title: "Mak ProSport" }, // Find by title
  { 
    $set: { 
      productId: ObjectId("PASTE_PRODUCT_ID_HERE") 
    } 
  }
)

// Verify
db.home_collection_features.findOne({ title: "Mak ProSport" })

// ===========================================
// EXAMPLE 3: List all products with their IDs
// ===========================================

db.products.find({}, { _id: 1, name: 1, brand: 1, price: 1 }).pretty()

// ===========================================
// EXAMPLE 4: List all hero slides
// ===========================================

db.hero_slides.find({}, { _id: 1, title: 1, subtitle: 1, productId: 1 }).pretty()

// ===========================================
// EXAMPLE 5: List all collection features
// ===========================================

db.home_collection_features.find({}, { _id: 1, title: 1, tagline: 1, productId: 1 }).pretty()

// ===========================================
// EXAMPLE 6: Remove product link from hero slide
// ===========================================

db.hero_slides.updateOne(
  { title: "MAK Chronograph" },
  { 
    $unset: { productId: "" } 
  }
)

// ===========================================
// EXAMPLE 7: Bulk update multiple slides
// ===========================================

// Array of updates
const updates = [
  {
    slideTitle: "MAK Chronograph",
    productName: "Chronograph Pro"
  },
  {
    slideTitle: "MAK Sport",
    productName: "Sport Elite"
  }
];

// Execute updates
updates.forEach(update => {
  // Find product
  const product = db.products.findOne({ name: new RegExp(update.productName, 'i') });
  
  if (product) {
    // Update slide
    db.hero_slides.updateOne(
      { title: update.slideTitle },
      { $set: { productId: product._id } }
    );
    print(`Linked "${update.slideTitle}" to product "${product.name}"`);
  } else {
    print(`Product "${update.productName}" not found`);
  }
});

// ===========================================
// EXAMPLE 8: Verify all links are working
// ===========================================

// Check hero slides with missing products
db.hero_slides.aggregate([
  {
    $lookup: {
      from: "products",
      localField: "productId",
      foreignField: "_id",
      as: "product"
    }
  },
  {
    $match: {
      productId: { $exists: true, $ne: null },
      product: { $size: 0 }
    }
  },
  {
    $project: {
      title: 1,
      productId: 1,
      message: "Product not found!"
    }
  }
])

// Check collection features with missing products
db.home_collection_features.aggregate([
  {
    $lookup: {
      from: "products",
      localField: "productId",
      foreignField: "_id",
      as: "product"
    }
  },
  {
    $match: {
      productId: { $exists: true, $ne: null },
      product: { $size: 0 }
    }
  },
  {
    $project: {
      title: 1,
      productId: 1,
      message: "Product not found!"
    }
  }
])

// ===========================================
// EXAMPLE 9: Get complete report of all links
// ===========================================

print("=== HERO SLIDES REPORT ===");
db.hero_slides.aggregate([
  {
    $lookup: {
      from: "products",
      localField: "productId",
      foreignField: "_id",
      as: "product"
    }
  },
  {
    $project: {
      title: 1,
      subtitle: 1,
      hasProductId: { $cond: [{ $gt: ["$productId", null] }, true, false] },
      productName: { $arrayElemAt: ["$product.name", 0] },
      productPrice: { $arrayElemAt: ["$product.price", 0] }
    }
  },
  {
    $sort: { position: 1 }
  }
]).forEach(printjson);

print("\n=== COLLECTION FEATURES REPORT ===");
db.home_collection_features.aggregate([
  {
    $lookup: {
      from: "products",
      localField: "productId",
      foreignField: "_id",
      as: "product"
    }
  },
  {
    $project: {
      title: 1,
      tagline: 1,
      hasProductId: { $cond: [{ $gt: ["$productId", null] }, true, false] },
      productName: { $arrayElemAt: ["$product.name", 0] },
      productPrice: { $arrayElemAt: ["$product.price", 0] }
    }
  },
  {
    $sort: { position: 1 }
  }
]).forEach(printjson);
