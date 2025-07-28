# Pehnaw API Implementation Guide

This document provides technical details about the implementation of the Pehnaw e-commerce API, aimed at developers who are extending or maintaining the codebase.

## System Architecture

The Pehnaw API is built with a clean architecture approach, organizing code into the following layers:

```
pehnaw-be/
├── cmd/               # Application entrypoints 
│   └── api/           # API server 
├── internal/          # Private application code
│   ├── config/        # Configuration management
│   ├── database/      # Database connections and operations
│   ├── handlers/      # Request handlers
│   ├── middleware/    # Middleware components
│   └── models/        # Data models
├── pkg/               # Public libraries
│   └── utils/         # Utility functions
```

### Key Components

1. **Entry Point**: `cmd/api/main.go`
   - Initializes configuration
   - Sets up database connections
   - Configures the web server
   - Handles graceful shutdown

2. **Configuration**: `internal/config/config.go`
   - Loads environment variables
   - Sets up MongoDB and Redis connections
   - Defines application settings

3. **Database Layer**: `internal/database/database.go`
   - Provides a unified client for both MongoDB and Redis
   - Implements CRUD operations for database entities
   - Handles caching with Redis

4. **API Handlers**: `internal/handlers/*.go`
   - Implements route handlers for different resource types
   - Validates incoming requests
   - Processes business logic
   - Returns formatted responses

5. **Middleware**: `internal/middleware/*.go`
   - JWT authentication
   - Role-based access control
   - Error handling

6. **Models**: `internal/models/*.go`
   - Defines data structures and validations
   - Provides request/response models

## Authentication Flow

1. User registration/login:
   - User submits credentials
   - Server validates credentials
   - On success, server generates JWT token with user ID, role, and expiration
   - Token is returned to the client

2. Authenticated requests:
   - Client includes token in `Authorization` header (`Bearer <token>`)
   - Auth middleware validates token signature and expiration
   - User information is extracted and stored in request context
   - Handlers can access user data via context

## Database Schema

### MongoDB Collections

1. **users**
   - `_id`: ObjectId (primary key)
   - `name`: String
   - `email`: String (indexed, unique)
   - `password`: String (hashed)
   - `role`: String
   - `created_at`: Timestamp
   - `updated_at`: Timestamp

2. **products**
   - `_id`: ObjectId (primary key)
   - `name`: String
   - `description`: String
   - `price`: Double
   - `category`: String (indexed)
   - `image_url`: String
   - `stock`: Integer
   - `created_at`: Timestamp
   - `updated_at`: Timestamp

3. **cart_items**
   - `_id`: ObjectId (primary key)
   - `user_id`: ObjectId (indexed)
   - `product_id`: ObjectId (indexed)
   - `quantity`: Integer
   - `created_at`: Timestamp
   - `updated_at`: Timestamp

4. **orders**
   - `_id`: ObjectId (primary key)
   - `user_id`: ObjectId (indexed)
   - `items`: Array of OrderItem objects
   - `total`: Double
   - `status`: String (indexed)
   - `shipping_address`: Address object
   - `payment_info`: PaymentInfo object
   - `created_at`: Timestamp (indexed)
   - `updated_at`: Timestamp

### Redis Cache Keys

- `products:{category}:{minPrice}:{maxPrice}:{sort}:{page}:{limit}` - Product listings
- `product:{id}` - Individual product details
- `cart:{userId}` - User's cart contents
- `orders:{userId}` - User's order history
- `recommendations:{userId}` - User's product recommendations

## Caching Strategy

1. **Cache-Aside Pattern**:
   - Check cache first
   - If cache miss, fetch from database
   - Update cache with fetched data

2. **Cache Invalidation**:
   - When data is modified, related cache entries are invalidated
   - Time-based expiration as a fallback

3. **Cache TTL (Time-to-Live)**:
   - Products: 10-30 minutes
   - Cart: 30 minutes
   - Orders: 15 minutes
   - Recommendations: 1 hour

## Error Handling

1. **Consistent Error Responses**:
   - All errors follow the same JSON format
   - HTTP status codes match error types

2. **Error Types**:
   - Validation errors (400)
   - Authentication errors (401)
   - Permission errors (403)
   - Not found errors (404)
   - Server errors (500)

## Security Considerations

1. **Password Security**:
   - Passwords are hashed using bcrypt
   - Sensitive data is never stored in plaintext

2. **JWT Security**:
   - Tokens are signed with a secret key
   - Tokens have an expiration time
   - Token validation checks signature and expiration

3. **Payment Information**:
   - CVV is never stored
   - Card numbers are partially masked before storage

## Performance Optimizations

1. **Indexing Strategy**:
   - Indexes on frequently queried fields (email, categories, etc.)
   - Compound indexes for complex queries

2. **Connection Pooling**:
   - MongoDB connections are pooled
   - Redis connections are reused

3. **Pagination**:
   - All list endpoints support pagination
   - Default limits prevent excessive data loading

4. **Efficient Queries**:
   - Projection to retrieve only needed fields
   - Use of aggregation pipeline for complex queries

## Adding New Endpoints

To add a new endpoint:

1. Create or modify a handler in `internal/handlers/`
2. Add the route in `internal/handlers/handlers.go`
3. Create any necessary models in `internal/models/`
4. Implement business logic in the handler
5. Add appropriate validation and error handling

Example:

```go
// 1. Create a handler function
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
    // Implementation
}

// 2. Add the route in SetupRoutes function
products.Post("/", middleware.Role("admin"), productHandler.CreateProduct)
```

## Testing

The codebase should be tested at multiple levels:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test API endpoints with a test database
3. **End-to-End Tests**: Test complete flows from client perspective

Test files should be placed next to the code they test with a `_test.go` suffix.

## Deployment Considerations

1. **Environment Variables**:
   - All configuration should be via environment variables
   - Use example.env as a template

2. **Database Migration**:
   - MongoDB schema evolves without formal migrations
   - Document changes should be backward compatible

3. **Monitoring**:
   - Log important events
   - Track API usage and performance

4. **Scaling**:
   - Application is stateless and can be horizontally scaled
   - Redis can be used for distributed locking if needed

## Common Tasks

### Adding a New Model

1. Create a new file in `internal/models/`
2. Define the struct with proper JSON and BSON tags
3. Add validation if needed

Example:

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
    ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
    CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}
```

### Adding a New Collection

1. Update the `Collections` method in `internal/database/database.go`

Example:

```go
// Collections returns MongoDB collections
func (db *DBClient) Collections() struct {
    Users        *mongo.Collection
    Products     *mongo.Collection
    CartItems    *mongo.Collection
    Orders       *mongo.Collection
    Categories   *mongo.Collection  // New collection
} {
    return struct {
        Users        *mongo.Collection
        Products     *mongo.Collection
        CartItems    *mongo.Collection
        Orders       *mongo.Collection
        Categories   *mongo.Collection  // New collection
    }{
        Users:        db.MongoDB.Collection("users"),
        Products:     db.MongoDB.Collection("products"),
        CartItems:    db.MongoDB.Collection("cart_items"),
        Orders:       db.MongoDB.Collection("orders"),
        Categories:   db.MongoDB.Collection("categories"),  // New collection
    }
}
```

### Adding a New Handler

1. Create a new file in `internal/handlers/` if needed
2. Implement the handler struct and methods
3. Register routes in `SetupRoutes`

Example:

```go
// category_handler.go
package handlers

// CategoryHandler handles category related requests
type CategoryHandler struct {
    DB     *database.DBClient
    Config *config.Config
}

// NewCategoryHandler creates a new instance of CategoryHandler
func NewCategoryHandler(db *database.DBClient, cfg *config.Config) *CategoryHandler {
    return &CategoryHandler{
        DB:     db,
        Config: cfg,
    }
}

// GetCategories returns all categories
func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
    // Implementation
}

// In handlers.go:
categoryHandler := NewCategoryHandler(db, cfg)
categories := app.Group("/categories")
categories.Get("/", categoryHandler.GetCategories)
```
