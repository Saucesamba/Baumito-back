package v1

import (
	"Avito-back/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"net/http"
)

type UserHandler struct {
	Usecase domain.UserUsecase
}

// Структура для приема данных из JSON
type registerInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"` // ДОБАВЬ ЭТУ СТРОКУ

}

func (h *UserHandler) Register(c *gin.Context) {
	var input registerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Email: input.Email,
		Name:  input.Name,
		Role:  input.Role, // ТЕПЕРЬ РОЛЬ ПЕРЕДАЕТСЯ В USECASE
	}

	if err := h.Usecase.Register(c.Request.Context(), user, input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully", "user_id": user.ID})
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.Usecase.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Возвращаем токен клиенту
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *UserHandler) BlockUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.Usecase.BlockUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user has been blocked"})
}
