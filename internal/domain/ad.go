package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type DealType string

const (
	DealTypeSell     DealType = "sell"
	DealTypeRent     DealType = "rent"
	DealTypeExchange DealType = "exchange"
	DealTypeFree     DealType = "free"
)

// Advertisement - модель товара в кампусе
type Ad struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	CategoryID   int                    `json:"category_id"`
	UniversityID int                    `json:"university_id"`
	LocationID   int                    `json:"location_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Price        float64                `json:"price"`
	Currency     string                 `json:"currency"`
	Status       string                 `json:"status"`
	DealType     DealType               `json:"deal_type"`
	ExtraProps   map[string]interface{} `json:"extra_props"` // Мапится на JSONB в Postgres
	ViewsCount   int                    `json:"views_count"`
	ExpiresAt    time.Time              `json:"expires_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type AdImage struct {
	ID       uuid.UUID `json:"id"`
	AdID     uuid.UUID `json:"ad_id"`
	ImageURL string    `json:"image_url"`
	IsMain   bool      `json:"is_main"`
}

type AdFilter struct {
	UniversityID int
	CategoryID   int
	SearchQuery  string
	Limit        int
	Offset       int
}

// AdRepository - контракт для БД
type AdRepository interface {
	Create(ctx context.Context, ad *Ad) error
	GetByID(ctx context.Context, id uuid.UUID) (*Ad, error)
	Update(ctx context.Context, ad *Ad) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddImage(ctx context.Context, img *AdImage) error
	Fetch(ctx context.Context, filter AdFilter) ([]*Ad, error)
}

// AdUsecase - правила создания и модерации
type AdUsecase interface {
	CreateAd(ctx context.Context, ad *Ad) error
	GetAd(ctx context.Context, id uuid.UUID) (*Ad, error)
	ListAds(ctx context.Context, filter AdFilter, page int) ([]*Ad, error)
	UploadImage(ctx context.Context, adID uuid.UUID, fileName string, file io.Reader, size int64) error
}
