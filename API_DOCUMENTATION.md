# Pehnaw API Documentation

This document provides comprehensive documentation for the Pehnaw e-commerce API. It covers all available endpoints, request/response formats, authentication requirements, and examples.

## Table of Contents
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Response Format](#response-format)
- [API Endpoints](#api-endpoints)
  - [Health and Welcome](#health-and-welcome)
  - [Authentication](#authentication-endpoints)
  - [Products](#products)
  - [Cart](#cart)
  - [Orders](#orders)
  - [Recommendations](#recommendations)
- [Data Models](#data-models)
- [Error Handling](#error-handling)

## Base URL

The base URL for all API endpoints is:

```
http://localhost:8080
```

This can be configured using the `PORT` environment variable.

## Authentication

Most endpoints in the API require authentication using JSON Web Tokens (JWT). To authenticate:

1. Obtain a token by registering or logging in
2. Include the token in the Authorization header of subsequent requests:

```
Authorization: Bearer <your_jwt_token>
```

## Response Format

All API responses follow a consistent structure:

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": { 
    // Response data specific to the endpoint
  }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

## API Endpoints

### Health and Welcome

#### GET /health
Check if the API is running properly.

**Authentication:** Not required

**Response:**
```json
{
  "success": true,
  "message": "Server is healthy"
}
```

#### GET /welcome
Get a welcome message from the API.

**Authentication:** Not required

**Response:**
```json
{
  "success": true,
  "message": "Welcome to Pehnaw API"
}
```

### Authentication Endpoints

#### POST /auth/register
Register a new user.

**Authentication:** Not required

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "60d21b4667d0d8992e610c85",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "user"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### POST /auth/login
Login to the system.

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "60d21b4667d0d8992e610c85",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "user",
      "picture": "https://example.com/profile.jpg", 
      "authProvider": "local"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### GET /auth/google
Initiates the Google OAuth login flow.

**Authentication:** Not required

**Response:** Redirects to Google's authentication page

#### GET /auth/google/callback
Handles the callback from Google's OAuth service.

**Authentication:** Not required

**Query Parameters:**
- `code` (string, required): Authorization code from Google
- `state` (string, required): State parameter for CSRF protection

**Response:** Redirects to the frontend with a token

#### GET /me
Get the current authenticated user's profile information.

**Authentication:** Required

**Response:**
```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "picture": "https://lh3.googleusercontent.com/a/example",
    "authProvider": "google"
  }
}
```

### Products

#### GET /products
Get a list of products with optional filtering.

**Authentication:** Not required

**Query Parameters:**
- `category` (string, optional): Filter products by category
- `minPrice` (number, optional): Minimum price filter
- `maxPrice` (number, optional): Maximum price filter
- `sortBy` (string, optional): Field to sort by (default: "createdAt")
- `order` (string, optional): Sort order - "asc" or "desc" (default: "desc")
- `page` (number, optional): Page number (default: 1)
- `limit` (number, optional): Results per page (default: 10)

**Response:**
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "60d21b4667d0d8992e610c85",
      "name": "Cotton T-Shirt",
      "description": "Comfortable cotton t-shirt",
      "price": 19.99,
      "category": "t-shirts",
      "imageUrl": "https://example.com/tshirt.jpg",
      "stock": 100,
      "createdAt": "2023-07-28T10:00:00Z",
      "updatedAt": "2023-07-28T10:00:00Z"
    },
    // More products...
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 50,
    "pages": 5
  }
}
```

#### GET /products/:id
Get details of a specific product.

**Authentication:** Not required

**URL Parameters:**
- `id` (string, required): Product ID

#### POST /products
Create a new product (admin only).

**Authentication:** Required (admin role)

**Request:** Multipart Form Data
- `name` (string, required): Product name
- `description` (string, required): Product description
- `price` (number, required): Product price
- `category` (string, required): Product category
- `stock` (number, required): Available stock
- `images` (files, optional): Multiple image files to upload

**Response:**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "name": "Cotton T-Shirt",
    "description": "Comfortable cotton t-shirt",
    "price": 19.99,
    "category": "t-shirts",
    "imageUrl": "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt.jpg",
    "images": [
      "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt.jpg",
      "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt-back.jpg"
    ],
    "stock": 100,
    "createdAt": "2023-07-28T10:00:00Z",
    "updatedAt": "2023-07-28T10:00:00Z"
  }
}
```

#### PUT /products/:id
Update an existing product (admin only).

**Authentication:** Required (admin role)

**URL Parameters:**
- `id` (string, required): Product ID

**Request:** Multipart Form Data
- `name` (string, optional): Product name
- `description` (string, optional): Product description
- `price` (number, optional): Product price
- `category` (string, optional): Product category
- `stock` (number, optional): Available stock
- `keepExistingImages` (boolean, optional): Whether to keep existing images (default: true)
- `images` (files, optional): New image files to upload

**Response:**
```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "name": "Premium Cotton T-Shirt",
    "description": "Comfortable premium cotton t-shirt",
    "price": 24.99,
    "category": "t-shirts",
    "imageUrl": "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt.jpg",
    "images": [
      "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt.jpg",
      "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt-back.jpg",
      "https://pehnaw.s3.ap-south-1.amazonaws.com/products/tshirt-side.jpg"
    ],
    "stock": 75,
    "createdAt": "2023-07-28T10:00:00Z",
    "updatedAt": "2023-07-29T15:30:00Z"
  }
}
```

#### DELETE /products/:id
Delete a product (admin only).

**Authentication:** Required (admin role)

**URL Parameters:**
- `id` (string, required): Product ID

**Query Parameters:**
- `deleteImages` (boolean, optional): Whether to delete images from S3 (default: false)

**Response:**
```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "name": "Cotton T-Shirt",
    "description": "Comfortable cotton t-shirt",
    "price": 19.99,
    "category": "t-shirts",
    "imageUrl": "https://example.com/tshirt.jpg",
    "stock": 100,
    "createdAt": "2023-07-28T10:00:00Z",
    "updatedAt": "2023-07-28T10:00:00Z"
  }
}
```

### Cart

#### POST /cart
Add a product to the cart.

**Authentication:** Required

**Request Body:**
```json
{
  "productId": "60d21b4667d0d8992e610c85",
  "quantity": 2
}
```

**Response:**
```json
{
  "success": true,
  "message": "Product added to cart successfully"
}
```

#### GET /cart/:userID
Get the current user's cart contents.

**Authentication:** Required

**URL Parameters:**
- `userID` (string, required): User ID (must match authenticated user or be an admin)

**Response:**
```json
{
  "success": true,
  "message": "Cart retrieved successfully",
  "data": {
    "items": [
      {
        "id": "60d21b4667d0d8992e610c85",
        "userId": "60d21b4667d0d8992e610c86",
        "productId": "60d21b4667d0d8992e610c87",
        "product": {
          "id": "60d21b4667d0d8992e610c87",
          "name": "Cotton T-Shirt",
          "description": "Comfortable cotton t-shirt",
          "price": 19.99,
          "category": "t-shirts",
          "imageUrl": "https://example.com/tshirt.jpg",
          "stock": 100,
          "createdAt": "2023-07-28T10:00:00Z",
          "updatedAt": "2023-07-28T10:00:00Z"
        },
        "quantity": 2,
        "createdAt": "2023-07-28T11:00:00Z",
        "updatedAt": "2023-07-28T11:00:00Z"
      }
      // More cart items...
    ],
    "total": 39.98
  }
}
```

#### DELETE /cart/:userID/:productID
Remove an item from the cart.

**Authentication:** Required

**URL Parameters:**
- `userID` (string, required): User ID (must match authenticated user or be an admin)
- `productID` (string, required): Product ID to remove

**Response:**
```json
{
  "success": true,
  "message": "Item removed from cart successfully"
}
```

### Orders

#### POST /checkout
Create a new order from cart items.

**Authentication:** Required

**Request Body:**
```json
{
  "shippingAddress": {
    "street": "123 Main Street",
    "city": "Anytown",
    "state": "State",
    "zipCode": "12345",
    "country": "Country"
  },
  "paymentInfo": {
    "method": "credit_card",
    "cardNumber": "4111111111111111",
    "expiryDate": "12/25",
    "cvv": "123"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Order placed successfully",
  "data": {
    "id": "60d21b4667d0d8992e610c88",
    "userId": "60d21b4667d0d8992e610c86",
    "items": [
      {
        "productId": "60d21b4667d0d8992e610c87",
        "productName": "Cotton T-Shirt",
        "price": 19.99,
        "quantity": 2,
        "subtotal": 39.98
      }
      // More order items...
    ],
    "total": 39.98,
    "status": "pending",
    "shippingAddress": {
      "street": "123 Main Street",
      "city": "Anytown",
      "state": "State",
      "zipCode": "12345",
      "country": "Country"
    },
    "paymentInfo": {
      "method": "credit_card",
      "cardNumber": "************1111",
      "expiryDate": "12/25"
    },
    "createdAt": "2023-07-28T12:00:00Z",
    "updatedAt": "2023-07-28T12:00:00Z"
  }
}
```

#### GET /orders/:userID
Get order history for a user.

**Authentication:** Required

**URL Parameters:**
- `userID` (string, required): User ID (must match authenticated user or be an admin)

**Response:**
```json
{
  "success": true,
  "message": "Orders retrieved successfully",
  "data": [
    {
      "id": "60d21b4667d0d8992e610c88",
      "userId": "60d21b4667d0d8992e610c86",
      "items": [
        {
          "productId": "60d21b4667d0d8992e610c87",
          "productName": "Cotton T-Shirt",
          "price": 19.99,
          "quantity": 2,
          "subtotal": 39.98
        }
      ],
      "total": 39.98,
      "status": "pending",
      "shippingAddress": {
        "street": "123 Main Street",
        "city": "Anytown",
        "state": "State",
        "zipCode": "12345",
        "country": "Country"
      },
      "paymentInfo": {
        "method": "credit_card",
        "cardNumber": "************1111",
        "expiryDate": "12/25"
      },
      "createdAt": "2023-07-28T12:00:00Z",
      "updatedAt": "2023-07-28T12:00:00Z"
    }
    // More orders...
  ]
}
```

### Recommendations

#### GET /recommendations/:userID
Get personalized product recommendations for a user.

**Authentication:** Required

**URL Parameters:**
- `userID` (string, required): User ID (must match authenticated user or be an admin)

**Response:**
```json
{
  "success": true,
  "message": "Product recommendations retrieved successfully",
  "data": [
    {
      "id": "60d21b4667d0d8992e610c89",
      "name": "Premium Hoodie",
      "description": "Warm premium hoodie",
      "price": 49.99,
      "category": "hoodies",
      "imageUrl": "https://example.com/hoodie.jpg",
      "stock": 50,
      "createdAt": "2023-07-28T10:00:00Z",
      "updatedAt": "2023-07-28T10:00:00Z"
    }
    // More product recommendations...
  ]
}
```

## Data Models

### User
```
{
  "id": "ObjectID",
  "name": "string",
  "email": "string",
  "password": "string (hashed, never returned in responses)",
  "role": "string (user, admin)",
  "googleId": "string (optional, for Google auth)",
  "picture": "string (optional, profile picture URL)",
  "authProvider": "string (local, google, hybrid)",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### Product
```
{
  "id": "ObjectID",
  "name": "string",
  "description": "string",
  "price": "float",
  "category": "string",
  "imageUrl": "string",
  "stock": "integer",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### Cart Item
```
{
  "id": "ObjectID",
  "userId": "ObjectID",
  "productId": "ObjectID",
  "product": "Product object (optional)",
  "quantity": "integer",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### Order
```
{
  "id": "ObjectID",
  "userId": "ObjectID",
  "items": [
    {
      "productId": "ObjectID",
      "productName": "string",
      "price": "float",
      "quantity": "integer",
      "subtotal": "float"
    }
  ],
  "total": "float",
  "status": "string (pending, processing, shipped, delivered, cancelled)",
  "shippingAddress": {
    "street": "string",
    "city": "string",
    "state": "string",
    "zipCode": "string",
    "country": "string"
  },
  "paymentInfo": {
    "method": "string",
    "cardNumber": "string (last 4 digits only)",
    "expiryDate": "string"
  },
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of requests:

- `200 OK`: Request successful
- `201 Created`: Resource successfully created
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Not enough permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error

Error responses include detailed error messages to help diagnose issues.

## Pagination

Endpoints that return multiple items (like `/products`) support pagination:

- `page`: Page number (starts at 1)
- `limit`: Number of results per page

Response includes pagination metadata:

```json
"meta": {
  "page": 1,
  "limit": 10,
  "total": 50,
  "pages": 5
}
```

## Filtering and Sorting

Product listing supports filtering by:
- Category
- Price range (min and max)

And sorting by:
- Any product field
- Ascending or descending order

## Caching

The API implements Redis caching for improved performance on:
- Product listings
- Individual product details
- Cart contents
- Order history
- Recommendations

Cache is automatically invalidated when data is modified.
