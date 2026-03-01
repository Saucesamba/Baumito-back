package v1

import (
	"net/http"

	"Avito-back/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	Usecase domain.ChatUsecase
}

// Структура для отправки сообщения
type sendMessageInput struct {
	AdID    string `json:"ad_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// 1. Отправка сообщения (POST /chats/messages)
func (h *ChatHandler) Send(c *gin.Context) {
	// Достаем ID отправителя из JWT
	senderIDStr, _ := c.Get("userId")
	senderID, _ := uuid.Parse(senderIDStr.(string))

	var input sendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adID, err := uuid.Parse(input.AdID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}

	if err := h.Usecase.SendMessage(c.Request.Context(), adID, senderID, input.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message sent"})
}

// 2. Получение истории сообщений (GET /chats/:id/messages)
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	userID, _ := uuid.Parse(userIDStr.(string))

	chatIDStr := c.Param("id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	msgs, err := h.Usecase.GetMessages(c.Request.Context(), chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": msgs})
}

func (h *ChatHandler) GetMyChats(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	userID, _ := uuid.Parse(userIDStr.(string))

	chats, err := h.Usecase.GetMyChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch chats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": chats})
}
