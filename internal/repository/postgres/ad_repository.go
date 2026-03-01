package postgres

import (
	"Avito-back/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type adRepository struct {
	storage *Storage
}

func NewAdRepository(storage *Storage) domain.AdRepository {
	return &adRepository{storage: storage}
}

func (r *adRepository) Create(ctx context.Context, ad *domain.Ad) error {
	query := `
		INSERT INTO advertisements (
			id, user_id, category_id, university_id, location_id, 
			title, description, price, currency, deal_type, extra_props
		) VALUES ($1, $2, $3, NULLIF($4, 0), NULLIF($5, 0), $6, $7, $8, $9, $10, $11)
	`
	if ad.ID == uuid.Nil {
		ad.ID = uuid.New()
	}

	_, err := r.storage.Pool.Exec(ctx, query,
		ad.ID, ad.UserID, ad.CategoryID, ad.UniversityID, ad.LocationID,
		ad.Title, ad.Description, ad.Price, ad.Currency, ad.DealType, ad.ExtraProps,
	)
	return err
}

func (r *adRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ad, error) {
	query := `SELECT id, user_id, title, description, price, deal_type, extra_props FROM advertisements WHERE id = $1`
	ad := &domain.Ad{}
	err := r.storage.Pool.QueryRow(ctx, query, id).Scan(
		&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.Price, &ad.DealType, &ad.ExtraProps,
	)
	return ad, err
}

// Остальные методы (Update, Delete) добавишь по аналогии
func (r *adRepository) Update(ctx context.Context, ad *domain.Ad) error { return nil }
func (r *adRepository) Delete(ctx context.Context, id uuid.UUID) error  { return nil }

func (r *adRepository) AddImage(ctx context.Context, img *domain.AdImage) error {
	query := `INSERT INTO ad_images (id, ad_id, image_url, is_main) VALUES ($1, $2, $3, $4)`
	if img.ID == uuid.Nil {
		img.ID = uuid.New()
	}
	_, err := r.storage.Pool.Exec(ctx, query, img.ID, img.AdID, img.ImageURL, img.IsMain)
	return err
}

func (r *adRepository) Fetch(ctx context.Context, filter domain.AdFilter) ([]*domain.Ad, error) {
	// Базовый запрос
	query := `
		SELECT id, user_id, title, price, deal_type, extra_props, created_at 
		FROM advertisements 
		WHERE status = 'active'
	`
	args := []interface{}{}
	argID := 1

	// 1. Фильтр по ВУЗу
	if filter.UniversityID != 0 {
		query += fmt.Sprintf(" AND university_id = $%d", argID)
		args = append(args, filter.UniversityID)
		argID++
	}

	// 2. Фильтр по Категории
	if filter.CategoryID != 0 {
		query += fmt.Sprintf(" AND category_id = $%d", argID)
		args = append(args, filter.CategoryID)
		argID++
	}

	// 3. ПОЛНОТЕКСТОВЫЙ ПОИСК (Самое крутое)
	if filter.SearchQuery != "" {
		// plainto_tsquery превращает "куплю учебник" в поисковый запрос для Postgres
		query += fmt.Sprintf(" AND search_vector @@ plainto_tsquery('russian', $%d)", argID)
		args = append(args, filter.SearchQuery)
		argID++
	}

	// 4. Пагинация
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.storage.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []*domain.Ad
	for rows.Next() {
		ad := &domain.Ad{}
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Price, &ad.DealType, &ad.ExtraProps, &ad.CreatedAt); err != nil {
			return nil, err
		}
		ads = append(ads, ad)
	}
	return ads, nil
}
