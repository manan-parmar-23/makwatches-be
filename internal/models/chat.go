package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessage represents a message in the chat support system
type ChatMessage struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ConversationID primitive.ObjectID `json:"conversationId" bson:"conversation_id"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	Content     string             `json:"content" bson:"content"`
	IsBot       bool               `json:"isBot" bson:"is_bot"`
	Timestamp   time.Time          `json:"timestamp" bson:"timestamp"`
}

// ChatConversation represents a chat conversation
type ChatConversation struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	Title       string             `json:"title" bson:"title"`
	Status      string             `json:"status" bson:"status"` // "active", "resolved", "archived"
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ChatMessageRequest is used for sending a message
type ChatMessageRequest struct {
	ConversationID string `json:"conversationId,omitempty"`
	Content        string `json:"content" validate:"required"`
}

// ChatConversationResponse represents a chat conversation with messages
type ChatConversationResponse struct {
	ID        primitive.ObjectID `json:"id"`
	UserID    primitive.ObjectID `json:"userId"`
	Title     string             `json:"title"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Messages  []ChatMessage      `json:"messages"`
}
