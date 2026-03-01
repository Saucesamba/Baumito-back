package v1

import (
	"Avito-back/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

type AdHandler struct {
	Usecase domain.AdUsecase
}

type createAdInput struct {
	Title        string                 `json:"title" binding:"required"`
	Description  string                 `json:"description" binding:"required"`
	Price        float64                `json:"price"`
	CategoryID   int                    `json:"category_id" binding:"required"`
	UniversityID int                    `json:"university_id" binding:"required"`
	LocationID   int                    `json:"location_id"`
	DealType     domain.DealType        `json:"deal_type" binding:"required"`
	ExtraProps   map[string]interface{} `json:"extra_props"`
}

func (h *AdHandler) Create(c *gin.Context) {
	// 1. Достаем ID пользователя из JWT (который проставил Middleware)
	userIDStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// 2. Валидируем входной JSON
	var input createAdInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad := &domain.Ad{
		UserID:       userID,
		Title:        input.Title,
		Description:  input.Description,
		Price:        input.Price,
		CategoryID:   input.CategoryID,
		UniversityID: input.UniversityID,
		LocationID:   input.LocationID,
		DealType:     input.DealType,
		ExtraProps:   input.ExtraProps,
	}

	if err := h.Usecase.CreateAd(c.Request.Context(), ad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ad"})
		return
	}

	c.JSON(http.StatusCreated, ad)
}

// 2. Получение одного объявления (Метод GET /ads/:id)
func (h *AdHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}

	ad, err := h.Usecase.GetAd(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ad not found"})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// 3. Получение списка объявлений (Метод GET /ads)
func (h *AdHandler) List(c *gin.Context) {
	// 1. Собираем параметры из запроса
	uniID, _ := strconv.Atoi(c.Query("university_id"))
	catID, _ := strconv.Atoi(c.Query("category_id"))
	search := c.Query("search") // Параметр ?search=учебник
	page, _ := strconv.Atoi(c.Query("page"))

	if page <= 0 {
		page = 1
	}

	// 2. Создаем структуру фильтра
	filter := domain.AdFilter{
		UniversityID: uniID,
		CategoryID:   catID,
		SearchQuery:  search,
	}

	// 3. Вызываем обновленный метод
	ads, err := h.Usecase.ListAds(c.Request.Context(), filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ads"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ads,
		"page": page,
	})
}

func (h *AdHandler) UploadImage(c *gin.Context) {
	adID, _ := uuid.Parse(c.Param("id"))

	// Читаем файл из multipart/form-data
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	if err := h.Usecase.UploadImage(c.Request.Context(), adID, header.Filename, file, header.Size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image uploaded successfully"})
}
