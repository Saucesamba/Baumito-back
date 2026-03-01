package chat

import (
	"Avito-back/internal/domain"
	"context"
	"github.com/google/uuid"
)

type chatUsecase struct {
	chatRepo domain.ChatRepository
	adRepo   domain.AdRepository
}

func NewChatUsecase(chatRepo domain.ChatRepository, adRepo domain.AdRepository) domain.ChatUsecase {
	return &chatUsecase{chatRepo: chatRepo, adRepo: adRepo}
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
	return u.chatRepo.CreateMessage(ctx, msg)
}

func (u *chatUsecase) GetMessages(ctx context.Context, chatID, userID uuid.UUID) ([]*domain.Message, error) {
	return u.chatRepo.GetChatMessages(ctx, chatID)
}

func (u *chatUsecase) GetMyChats(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error) {
	return u.chatRepo.GetUserChats(ctx, userID)
}
