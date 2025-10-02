# Firebase Storage Setup for MakWatches Backend

This guide explains how to configure Firebase Storage for image uploads in the MakWatches backend.

## Prerequisites

1. A Firebase project (✅ Already created: makwatches-1ae1a)
2. Firebase Storage enabled in your Firebase project (❌ **NEEDS TO BE DONE**)
3. A service account with appropriate permissions (✅ Already configured)

## Setup Steps

### 1. Enable Firebase Storage (REQUIRED)

**IMPORTANT**: You need to enable Firebase Storage in your project first!

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project **makwatches-1ae1a**
3. In the left sidebar, click on **"Storage"**
4. Click **"Get started"**
5. Choose security rules:
   - For development: Select "Start in test mode"
   - For production: Configure proper security rules
6. Select a location for your default bucket (recommend: us-central1 or asia-south1)
7. Click **"Done"**

After enabling, your bucket will be available at: `makwatches-1ae1a.appspot.com`

### 2. Create a Service Account

1. Go to Project Settings > Service Accounts
2. Click "Generate new private key"
3. Download the JSON file
4. Rename it to `firebase-admin.json` and place it in the root of your project

### 3. Configure Environment Variables

Add the following to your `.env` file:

```env
# Firebase Configuration
FIREBASE_CREDENTIALS_PATH=firebase-admin.json
FIREBASE_BUCKET_NAME=your-project-id.appspot.com
```

Replace `your-project-id` with your actual Firebase project ID.

### 4. Security Rules

Update your Firebase Storage security rules to allow public read access:

```javascript
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /{allPaths=**} {
      allow read: if true;
      allow write: if request.auth != null;
    }
  }
}
```

## Usage

Once configured, the backend will automatically:

1. Upload images to Firebase Storage when creating/updating products
2. Generate public URLs for the uploaded images
3. Store these URLs in MongoDB
4. Return Firebase Storage URLs in API responses

## File Structure

```
your-project/
├── firebase-admin.json          # Service account credentials (DO NOT COMMIT)
├── firebase-admin.json.example  # Example credentials file
├── .env                         # Environment variables (DO NOT COMMIT)
└── example.env                  # Example environment file
```

## Important Security Notes

- **Never commit** `firebase-admin.json` to version control
- **Never commit** `.env` files with real credentials
- Use environment variables in production
- Regularly rotate service account keys
- Monitor Firebase Storage usage and costs

## Testing

You can test the upload functionality using:

1. Admin panel (if available)
2. Direct API calls to `/upload` endpoint (requires admin authentication)
3. Product creation endpoints that include image uploads

## Troubleshooting

### Common Issues

1. **"Failed to initialize Firebase client"**

   - Check if `firebase-admin.json` exists and has correct permissions
   - Verify the service account has Storage Admin role

2. **"Failed to upload to Firebase Storage"**

   - Check Firebase Storage rules
   - Verify bucket name is correct
   - Ensure Firebase Storage is enabled in your project

3. **"Permission denied"**
   - Check service account permissions
   - Verify Storage Admin role is assigned
   - Check Firebase Storage security rules

### Debug Steps

1. Check logs for detailed error messages
2. Verify environment variables are loaded correctly
3. Test Firebase credentials manually
4. Check Firebase project settings and quotas
