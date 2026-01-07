# Pickup Location Fix for Delhivery Integration

## Issue
The pickup location was not being properly displayed in orders because the seller/pickup address was not being sent correctly to the Delhivery API.

## Fixed Pickup Location
**Shree Ganesh Watch, Matva Street, Near Balaji Complex, Stand chowk, Jetpur, Rajkot**

## Changes Made

### 1. Updated `internal/services/delhivery.go`
- Added fallback logic to ensure pickup address is always set
- Added detailed logging to track what pickup information is being sent to Delhivery
- The shipment now explicitly includes:
  - `SellerName`: Mak Watches
  - `SellerAdd`: Shree Ganesh Watch, Matva Street, Near Balaji Complex, Stand chowk, Jetpur, Rajkot
  - `SellerCity`: Jetpur
  - `SellerState`: Gujarat
  - `SellerPin`: 360370
  - `SellerPhone`: 9974959693

### 2. Updated `internal/config/config.go`
- Added default fallback values for all Delhivery seller/pickup configuration
- Ensures that even if environment variables are not set, the correct pickup location will be used

### 3. Environment Variables (Already Configured)
The `.env` file already has the correct configuration:
```env
DELHIVERY_PICKUP_LOCATION=Shree Ganesh Watch
DELHIVERY_SELLER_NAME=Mak Watches
DELHIVERY_SELLER_PHONE=9974959693
DELHIVERY_SELLER_ADDRESS=Shree Ganesh Watch, Matva Street, Near Balaji Complex, Stand chowk, Jetpur, Rajkot
DELHIVERY_SELLER_CITY=Jetpur
DELHIVERY_SELLER_STATE=Gujarat
DELHIVERY_SELLER_PINCODE=360370
```

## How It Works

When a shipment is created with Delhivery:

1. **pickup_location Parameter**: Sent as `"Shree Ganesh Watch"` in the form data
2. **Seller/Pickup Fields in JSON**: The shipment JSON includes complete seller address details:
   - `seller_name`: "Mak Watches"
   - `seller_add`: "Shree Ganesh Watch, Matva Street, Near Balaji Complex, Stand chowk, Jetpur, Rajkot"
   - `seller_city`: "Jetpur"
   - `seller_state`: "Gujarat"
   - `seller_pin`: "360370"
   - `seller_phone`: "9974959693"

## Verification

After deploying these changes:

1. Place a new test order
2. Check the server logs for lines starting with `[DELHIVERY-API]`
3. You should see detailed pickup/seller information being logged
4. The order tracking page should now show the complete pickup address instead of "null"

## Testing

To test locally:
```bash
# Restart the backend server
go run main.go
```

To deploy to production:
```bash
# Rebuild and redeploy the Docker container
docker-compose down
docker-compose up -d --build
```

## Important Notes

1. The pickup location name (`Shree Ganesh Watch`) must be registered with Delhivery
2. The address must match what you registered with Delhivery during account setup
3. If you see warnings in logs about empty environment variables, check that your `.env` file is being loaded correctly
4. The API will use fallback values if environment variables are missing, ensuring the pickup location is always sent

## Logs to Look For

When creating a shipment, you'll now see detailed logs:
```
[DELHIVERY-API] Pickup/Seller Details Being Sent:
  - Name: Mak Watches
  - Address: Shree Ganesh Watch, Matva Street, Near Balaji Complex, Stand chowk, Jetpur, Rajkot
  - City: Jetpur
  - State: Gujarat
  - Pincode: 360370
  - Phone: 9974959693
[DELHIVERY-API] ✓ Form data - pickup_location parameter: 'Shree Ganesh Watch'
```

This confirms that all pickup information is being sent to Delhivery correctly.
