package firebase

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// FirebaseClient wraps the GCS client for Firebase Storage
// Usage: client, err := NewFirebaseClient(ctx, "path/to/serviceAccountKey.json", "your-bucket-name")
type FirebaseClient struct {
	StorageClient *storage.Client
	BucketName    string
}

func NewFirebaseClient(ctx context.Context, credentialsPath, bucketName string) (*FirebaseClient, error) {
	log.Printf("[FIREBASE] Initializing client with credentials: %s, bucket: %s", credentialsPath, bucketName)
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Printf("[FIREBASE] Failed to create storage client: %v", err)
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}
	log.Println("[FIREBASE] Storage client created successfully")

	// Validate bucket exists
	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
	if err != nil {
		log.Printf("[FIREBASE] Bucket validation failed: %v", err)
		if contains404(err.Error()) {
			return nil, fmt.Errorf("Firebase Storage bucket %q does not exist. Please verify the bucket name and enable Firebase Storage in your Firebase Console (https://console.firebase.google.com/).")
		}
		return nil, fmt.Errorf("failed to access bucket: %w", err)
	}
	log.Printf("[FIREBASE] Bucket %s validated successfully", bucketName)

	return &FirebaseClient{
		StorageClient: client,
		BucketName:    bucketName,
	}, nil
}

func contains404(errStr string) bool {
	return strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "bucket doesn't exist") ||
		strings.Contains(errStr, "bucket does not exist")
}

// UploadFile uploads a file to Firebase Storage and returns the public URL
func (f *FirebaseClient) UploadFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(filename))
	log.Printf("[FIREBASE] Starting upload for %s as %s to bucket %s", filename, objectName, f.BucketName)

	wc := f.StorageClient.Bucket(f.BucketName).Object(objectName).NewWriter(ctx)

	// Detect content type from file extension
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		wc.ContentType = "image/jpeg"
	case ".png":
		wc.ContentType = "image/png"
	case ".gif":
		wc.ContentType = "image/gif"
	case ".webp":
		wc.ContentType = "image/webp"
	default:
		wc.ContentType = "image/jpeg" // default
	}
	log.Printf("[FIREBASE] Set content type: %s", wc.ContentType)

	log.Println("[FIREBASE] Copying file data...")
	if _, err := io.Copy(wc, file); err != nil {
		log.Printf("[FIREBASE] Failed to copy file data: %v", err)
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	log.Println("[FIREBASE] Closing writer...")
	if err := wc.Close(); err != nil {
		log.Printf("[FIREBASE] Failed to close writer: %v", err)
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	log.Println("[FIREBASE] Setting public access...")
	// Make the file public
	obj := f.StorageClient.Bucket(f.BucketName).Object(objectName)
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		log.Printf("[FIREBASE] Failed to set public access: %v", err)
		return "", fmt.Errorf("failed to set public access: %w", err)
	}

	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", f.BucketName, objectName)
	log.Printf("[FIREBASE] Upload completed successfully, URL: %s", publicURL)
	return publicURL, nil
}
