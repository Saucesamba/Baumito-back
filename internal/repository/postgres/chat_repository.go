package postgres

import (
	"Avito-back/internal/domain"
	"context"
	"github.com/google/uuid"
)

type chatRepository struct {
	storage *Storage
}

func NewChatRepository(storage *Storage) domain.ChatRepository {
	return &chatRepository{storage: storage}
}

func (r *chatRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	query := `INSERT INTO chats (id, ad_id, buyer_id, seller_id) VALUES ($1, $2, $3, $4)`
	if chat.ID == uuid.Nil {
		chat.ID = uuid.New()
	}
	_, err := r.storage.Pool.Exec(ctx, query, chat.ID, chat.AdID, chat.BuyerID, chat.SellerID)
	return err
}

func (r *chatRepository) GetChatByParticipants(ctx context.Context, adID, buyerID, sellerID uuid.UUID) (*domain.Chat, error) {
	query := `SELECT id, ad_id, buyer_id, seller_id FROM chats WHERE ad_id = $1 AND buyer_id = $2 AND seller_id = $3`
	chat := &domain.Chat{}
	err := r.storage.Pool.QueryRow(ctx, query, adID, buyerID, sellerID).Scan(&chat.ID, &chat.AdID, &chat.BuyerID, &chat.SellerID)
	return chat, err
}

func (r *chatRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	query := `INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4)`
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	_, err := r.storage.Pool.Exec(ctx, query, msg.ID, msg.ChatID, msg.SenderID, msg.Content)
	return err
}

func (r *chatRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*domain.Message, error) {
	query := `SELECT id, chat_id, sender_id, content, is_read, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at ASC`
	rows, err := r.storage.Pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs := make([]*domain.Message, 0)
	for rows.Next() {
		m := &domain.Message{}
		rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Content, &m.IsRead, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *chatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error) {
	query := `
		SELECT id, ad_id, buyer_id, seller_id, created_at 
		FROM chats 
		WHERE buyer_id = $1 OR seller_id = $1 
		ORDER BY created_at DESC;`
	rows, err := r.storage.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]*domain.Chat, 0)
	for rows.Next() {
		c := &domain.Chat{}
		rows.Scan(&c.ID, &c.AdID, &c.BuyerID, &c.SellerID, &c.CreatedAt)
		chats = append(chats, c)
	}
	return chats, nil
}
