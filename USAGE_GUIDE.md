# Pehnaw API Usage Guide

This guide provides practical examples of how to use the Pehnaw e-commerce API. It includes code snippets for common operations across different platforms.

## Table of Contents
- [Setup](#setup)
- [Authentication](#authentication)
- [Working with Products](#working-with-products)
- [Shopping Cart Operations](#shopping-cart-operations)
- [Checkout Process](#checkout-process)
- [Order Management](#order-management)
- [Recommendations](#recommendations)
- [Error Handling](#error-handling)

## Setup

Before using the API, make sure the server is running properly. You can check this with a simple health check:

```javascript
// JavaScript (Fetch API)
fetch('http://localhost:8080/health')
  .then(response => response.json())
  .then(data => console.log(data));
```

```python
# Python (Requests)
import requests
response = requests.get('http://localhost:8080/health')
print(response.json())
```

```go
// Go
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
)

func main() {
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println(string(body))
}
```

## Authentication

### User Registration

To register a new user:

```javascript
// JavaScript
fetch('http://localhost:8080/auth/register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'John Doe',
    email: 'john@example.com',
    password: 'securepassword'
  }),
})
.then(response => response.json())
.then(data => {
  console.log(data);
  // Save the token
  localStorage.setItem('token', data.data.token);
});
```

```python
# Python
import requests

response = requests.post(
    'http://localhost:8080/auth/register',
    json={
        'name': 'John Doe',
        'email': 'john@example.com',
        'password': 'securepassword'
    }
)
data = response.json()
print(data)
# Save the token
token = data['data']['token']
```

### User Login

To login with existing credentials:

```javascript
// JavaScript
fetch('http://localhost:8080/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'john@example.com',
    password: 'securepassword'
  }),
})
.then(response => response.json())
.then(data => {
  console.log(data);
  // Save the token
  localStorage.setItem('token', data.data.token);
});
```

```python
# Python
import requests

response = requests.post(
    'http://localhost:8080/auth/login',
    json={
        'email': 'john@example.com',
        'password': 'securepassword'
    }
)
data = response.json()
print(data)
# Save the token
token = data['data']['token']
```

## Working with Products

### Browsing Products

To get a list of products with filtering:

```javascript
// JavaScript
fetch('http://localhost:8080/products?category=t-shirts&minPrice=10&maxPrice=50&page=1&limit=10')
  .then(response => response.json())
  .then(data => console.log(data));
```

```python
# Python
import requests

params = {
    'category': 't-shirts',
    'minPrice': 10,
    'maxPrice': 50,
    'page': 1,
    'limit': 10
}
response = requests.get('http://localhost:8080/products', params=params)
print(response.json())
```

### Getting Product Details

To get details of a specific product:

```javascript
// JavaScript
const productId = '60d21b4667d0d8992e610c85';
fetch(`http://localhost:8080/products/${productId}`)
  .then(response => response.json())
  .then(data => console.log(data));
```

```python
# Python
import requests

product_id = '60d21b4667d0d8992e610c85'
response = requests.get(f'http://localhost:8080/products/{product_id}')
print(response.json())
```

## Shopping Cart Operations

### Adding Products to Cart

To add a product to the cart:

```javascript
// JavaScript
const token = localStorage.getItem('token');
fetch('http://localhost:8080/cart', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    productId: '60d21b4667d0d8992e610c85',
    quantity: 2
  }),
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
response = requests.post(
    'http://localhost:8080/cart',
    headers={'Authorization': f'Bearer {token}'},
    json={
        'productId': '60d21b4667d0d8992e610c85',
        'quantity': 2
    }
)
print(response.json())
```

### Viewing Cart Contents

To view the current cart:

```javascript
// JavaScript
const token = localStorage.getItem('token');
const userId = 'your_user_id'; // Get this from token or user data
fetch(`http://localhost:8080/cart/${userId}`, {
  headers: {
    'Authorization': `Bearer ${token}`
  },
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
user_id = "your_user_id"  # Get this from token or user data
response = requests.get(
    f'http://localhost:8080/cart/{user_id}',
    headers={'Authorization': f'Bearer {token}'}
)
print(response.json())
```

### Removing Items from Cart

To remove an item from the cart:

```javascript
// JavaScript
const token = localStorage.getItem('token');
const userId = 'your_user_id';
const productId = '60d21b4667d0d8992e610c85';
fetch(`http://localhost:8080/cart/${userId}/${productId}`, {
  method: 'DELETE',
  headers: {
    'Authorization': `Bearer ${token}`
  },
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
user_id = "your_user_id"
product_id = "60d21b4667d0d8992e610c85"
response = requests.delete(
    f'http://localhost:8080/cart/{user_id}/{product_id}',
    headers={'Authorization': f'Bearer {token}'}
)
print(response.json())
```

## Checkout Process

To place an order:

```javascript
// JavaScript
const token = localStorage.getItem('token');
fetch('http://localhost:8080/checkout', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    shippingAddress: {
      street: '123 Main Street',
      city: 'Anytown',
      state: 'State',
      zipCode: '12345',
      country: 'Country'
    },
    paymentInfo: {
      method: 'credit_card',
      cardNumber: '4111111111111111',
      expiryDate: '12/25',
      cvv: '123'
    }
  }),
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
response = requests.post(
    'http://localhost:8080/checkout',
    headers={'Authorization': f'Bearer {token}'},
    json={
        'shippingAddress': {
            'street': '123 Main Street',
            'city': 'Anytown',
            'state': 'State',
            'zipCode': '12345',
            'country': 'Country'
        },
        'paymentInfo': {
            'method': 'credit_card',
            'cardNumber': '4111111111111111',
            'expiryDate': '12/25',
            'cvv': '123'
        }
    }
)
print(response.json())
```

## Order Management

### Viewing Order History

To view a user's order history:

```javascript
// JavaScript
const token = localStorage.getItem('token');
const userId = 'your_user_id';
fetch(`http://localhost:8080/orders/${userId}`, {
  headers: {
    'Authorization': `Bearer ${token}`
  },
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
user_id = "your_user_id"
response = requests.get(
    f'http://localhost:8080/orders/{user_id}',
    headers={'Authorization': f'Bearer {token}'}
)
print(response.json())
```

## Recommendations

To get personalized product recommendations:

```javascript
// JavaScript
const token = localStorage.getItem('token');
const userId = 'your_user_id';
fetch(`http://localhost:8080/recommendations/${userId}`, {
  headers: {
    'Authorization': `Bearer ${token}`
  },
})
.then(response => response.json())
.then(data => console.log(data));
```

```python
# Python
import requests

token = "your_saved_token"
user_id = "your_user_id"
response = requests.get(
    f'http://localhost:8080/recommendations/{user_id}',
    headers={'Authorization': f'Bearer {token}'}
)
print(response.json())
```

## Error Handling

Always check the response status and handle errors appropriately:

```javascript
// JavaScript
fetch('http://localhost:8080/products/invalid_id')
  .then(response => {
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    return response.json();
  })
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));
```

```python
# Python
import requests

try:
    response = requests.get('http://localhost:8080/products/invalid_id')
    response.raise_for_status()  # Raises an exception for 4XX/5XX responses
    data = response.json()
    print(data)
except requests.exceptions.HTTPError as error:
    print(f"HTTP Error: {error}")
    if response.text:
        print(f"Error details: {response.json()}")
except Exception as error:
    print(f"Error: {error}")
```

## Helper Function for API Calls

Here's a reusable function for JavaScript that handles API calls with proper error handling:

```javascript
// Helper function for API calls
async function callApi(endpoint, method = 'GET', data = null, token = null) {
  const headers = {
    'Content-Type': 'application/json',
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  const options = {
    method,
    headers,
  };
  
  if (data && (method === 'POST' || method === 'PUT')) {
    options.body = JSON.stringify(data);
  }
  
  try {
    const response = await fetch(`http://localhost:8080${endpoint}`, options);
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(result.message || 'API error');
    }
    
    return result;
  } catch (error) {
    console.error('API call failed:', error);
    throw error;
  }
}

// Usage examples:
// Get products
callApi('/products?category=t-shirts')
  .then(data => console.log(data))
  .catch(error => console.error(error));

// Login
callApi('/auth/login', 'POST', { email: 'user@example.com', password: 'password' })
  .then(data => console.log(data))
  .catch(error => console.error(error));

// Add to cart (with authentication)
const token = localStorage.getItem('token');
callApi('/cart', 'POST', { productId: 'product_id', quantity: 1 }, token)
  .then(data => console.log(data))
  .catch(error => console.error(error));
```
