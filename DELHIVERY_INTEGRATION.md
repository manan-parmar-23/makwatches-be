# Delhivery Shipping Integration Guide

This document explains the Delhivery One shipping integration for automated order fulfillment.

## Overview

When a customer places an order, the system automatically:
1. Creates a shipment with Delhivery
2. Gets a waybill (tracking number)
3. Updates the order with tracking information
4. Receives webhook updates for delivery status changes

## Configuration

### Environment Variables

Add these to your `.env` file:

```env
# Delhivery Configuration
DELHIVERY_API_TOKEN=your_api_token_here
DELHIVERY_BASE_URL=https://track.delhivery.com
DELHIVERY_PICKUP_LOCATION=your_registered_location_name

# Seller/Pickup Details
DELHIVERY_SELLER_NAME=Mak Watches
DELHIVERY_SELLER_PHONE=9876543210
DELHIVERY_SELLER_ADDRESS=Your Business Address
DELHIVERY_SELLER_CITY=Mumbai
DELHIVERY_SELLER_STATE=Maharashtra
DELHIVERY_SELLER_PINCODE=400001

# Return Address
DELHIVERY_RETURN_ADDRESS=Your Return Address
DELHIVERY_RETURN_CITY=Mumbai
DELHIVERY_RETURN_STATE=Maharashtra
DELHIVERY_RETURN_PINCODE=400001
DELHIVERY_RETURN_PHONE=9876543210
```

### Important Notes

1. **API Token**: Get this from your Delhivery One dashboard
2. **Base URL**: 
   - Production: `https://track.delhivery.com`
   - Staging/Testing: `https://staging-express.delhivery.com`
3. **Pickup Location**: Must be registered with Delhivery beforehand

## Order Flow

### 1. Order Placement

When a customer completes checkout:

```
Customer → Checkout → Order Created → Delhivery Shipment Created (async)
```

The shipment creation happens in the background so the customer doesn't wait.

### 2. Payment Modes

The system automatically determines payment mode:

| Payment Method | Delhivery Mode | COD Amount |
|---------------|----------------|------------|
| Razorpay/UPI/Card | Prepaid | ₹0 |
| Cash on Delivery | COD | Order Total |

### 3. Tracking

Order tracking info is stored in the `shippingInfo` field:

```json
{
  "shippingInfo": {
    "provider": "delhivery",
    "waybill": "1234567890",
    "trackingUrl": "https://www.delhivery.com/track/package/1234567890",
    "shipmentStatus": "manifested",
    "expectedDelivery": "2024-12-05",
    "currentLocation": "Mumbai Hub"
  }
}
```

## API Endpoints

### Public Endpoints

#### Check Pincode Serviceability
```
GET /shipping/check-pincode/:pincode
GET /shipping/check-pincode?pincode=400001
```

Response:
```json
{
  "success": true,
  "serviceable": true,
  "data": {
    "pincode": "400001",
    "city": "Mumbai",
    "state": "Maharashtra",
    "cod": true,
    "prepaid": true
  }
}
```

#### Track by Waybill
```
GET /shipping/track/:waybill
```

### Authenticated Endpoints

#### Track Order Shipment
```
GET /shipping/track/order/:orderID
Authorization: Bearer <token>
```

### Admin Endpoints

#### Retry Failed Shipment
```
POST /admin/shipping/orders/:orderID/retry
```

#### Cancel Shipment
```
POST /admin/shipping/orders/:orderID/cancel
```

#### Get Shipping Label
```
GET /admin/shipping/orders/:orderID/label
```

#### Bulk Track Shipments
```
POST /admin/shipping/bulk-track
Body: { "waybills": ["1234567890", "0987654321"] }
```

#### Request Pickup
```
POST /admin/shipping/request-pickup
Body: { "pickupDate": "2024-12-01", "expectedPackages": 5 }
```

### Webhook Endpoint

Delhivery sends status updates to:
```
POST /webhooks/delhivery
```

Configure this URL in your Delhivery dashboard.

## Status Mapping

| Delhivery Status | Order Status |
|------------------|--------------|
| Manifested (MD) | processing |
| Picked Up (PU) | shipped |
| In Transit (IT/OT) | shipped |
| Out for Delivery (OD) | out_for_delivery |
| Delivered (DL) | delivered |
| Undelivered (UD) | shipped (with note) |
| RTO (RT) | returned |

## Error Handling

If shipment creation fails:
1. Error is logged
2. Order is saved with `shippingInfo.shipmentError`
3. Admin can retry via `/admin/shipping/orders/:orderID/retry`

## Webhook Setup

1. Go to Delhivery One dashboard
2. Navigate to Webhook Settings
3. Add webhook URL: `https://your-domain.com/webhooks/delhivery`
4. Select events: All status updates

## Package Dimensions

Default package dimensions for watches:
- Weight: 500g
- Length: 15cm
- Breadth: 10cm
- Height: 8cm

Modify in `order_handler.go` if needed.

## Testing

For testing, use:
- Staging URL: `https://staging-express.delhivery.com`
- Test pincodes provided by Delhivery

## Troubleshooting

### Shipment Not Created

1. Check logs for errors
2. Verify Delhivery API token is valid
3. Ensure pickup location is registered
4. Verify customer pincode is serviceable

### Webhook Not Received

1. Verify webhook URL is accessible publicly
2. Check Delhivery webhook logs
3. Ensure HTTPS is properly configured

### Tracking Not Updating

1. Wait for Delhivery to process the shipment
2. Try manual refresh via tracking endpoint
3. Check if webhook is configured correctly

## Support

For Delhivery API issues:
- Documentation: https://www.delhivery.com/developers
- Support: support@delhivery.com
