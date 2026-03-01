package chat

import (
	"Avito-back/internal/domain"
	"Avito-back/internal/repository/kafka"
	"context"
	"log"

	"github.com/google/uuid"
)

type chatUsecase struct {
	chatRepo domain.ChatRepository
	adRepo   domain.AdRepository
	notifier *kafka.NotificationProducer // Добавь это поле

}

func NewChatUsecase(cr domain.ChatRepository, ar domain.AdRepository, n *kafka.NotificationProducer) domain.ChatUsecase {
	return &chatUsecase{chatRepo: cr, adRepo: ar, notifier: n}
}

func (u *chatUsecase) SendMessage(ctx context.Context, adID, senderID uuid.UUID, content string) error {
	// 1. Находим объявление, чтобы узнать, кто продавец
	ad, err := u.adRepo.GetByID(ctx, adID)
	if err != nil {
		return err
	}

	// 2. Ищем существующий чат между этим покупателем и продавцом по этому товару
	chat, err := u.chatRepo.GetChatByParticipants(ctx, adID, senderID, ad.UserID)
	if err != nil {
		// Если чата нет - создаем новый
		chat = &domain.Chat{ID: uuid.New(), AdID: adID, BuyerID: senderID, SellerID: ad.UserID}
		if err := u.chatRepo.CreateChat(ctx, chat); err != nil {
			return err
		}
	}

	// 3. Сохраняем сообщение
	msg := &domain.Message{ChatID: chat.ID, SenderID: senderID, Content: content}
	if err := u.chatRepo.CreateMessage(ctx, msg); err != nil {
		return err
	}

	// АСИНХРОННАЯ ЧАСТЬ: Кидаем событие в Kafka

	go func() {
		err = u.notifier.PublishMessageEvent(context.Background(), map[string]interface{}{
			"type":    "new_message",
			"chat_id": chat.ID,
			"sender":  senderID,
			"text":    content,
		})
		if err != nil {
			log.Printf("Failed to push to Kafka: %v", err)
		}
	}()
	return nil

}

func (u *chatUsecase) GetMessages(ctx context.Context, chatID, userID uuid.UUID) ([]*domain.Message, error) {
	return u.chatRepo.GetChatMessages(ctx, chatID)
}

func (u *chatUsecase) GetMyChats(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error) {
	return u.chatRepo.GetUserChats(ctx, userID)
}
