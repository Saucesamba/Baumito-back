package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID `json:"id"`
	AdID      uuid.UUID `json:"ad_id"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	ChatID    uuid.UUID `json:"chat_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository interface {
	CreateChat(ctx context.Context, chat *Chat) error
	GetChatByParticipants(ctx context.Context, adID, buyerID, sellerID uuid.UUID) (*Chat, error)
	CreateMessage(ctx context.Context, msg *Message) error
	GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*Message, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]*Chat, error)
}

type ChatUsecase interface {
	SendMessage(ctx context.Context, adID, senderID uuid.UUID, content string) error
	GetMessages(ctx context.Context, chatID, userID uuid.UUID) ([]*Message, error)
	GetMyChats(ctx context.Context, userID uuid.UUID) ([]*Chat, error)
}
