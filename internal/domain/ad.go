package domain

import (
	"context"
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
type Advertisement struct {
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

// AdRepository - контракт для БД
type AdRepository interface {
	Create(ctx context.Context, ad *Advertisement) error
	GetByID(ctx context.Context, id uuid.UUID) (*Advertisement, error)
	Update(ctx context.Context, ad *Advertisement) error
	Delete(ctx context.Context, id uuid.UUID) error
	// Fetch - для листинга с фильтрами (вуз, категория, цена)
	Fetch(ctx context.Context, universityID int, categoryID int, limit, offset int) ([]*Advertisement, error)
}

// AdUsecase - правила создания и модерации
type AdUsecase interface {
	CreateAd(ctx context.Context, ad *Advertisement) error
	GetAd(ctx context.Context, id uuid.UUID) (*Advertisement, error)
	ListAds(ctx context.Context, universityID int, categoryID int, page int) ([]*Advertisement, error)
}
