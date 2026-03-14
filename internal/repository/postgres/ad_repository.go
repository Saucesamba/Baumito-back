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
	// Выбираем все важные поля, включая university_id и location_id
	query := `
		SELECT 
			id, user_id, category_id, university_id, location_id, 
			title, description, price, currency, status, deal_type, 
			extra_props, views_count, expires_at, created_at, updated_at 
		FROM advertisements 
		WHERE id = $1
	`

	ad := &domain.Ad{}
	err := r.storage.Pool.QueryRow(ctx, query, id).Scan(
		&ad.ID, &ad.UserID, &ad.CategoryID, &ad.UniversityID, &ad.LocationID,
		&ad.Title, &ad.Description, &ad.Price, &ad.Currency, &ad.Status, &ad.DealType,
		&ad.ExtraProps, &ad.ViewsCount, &ad.ExpiresAt, &ad.CreatedAt, &ad.UpdatedAt,
	)

	if err != nil {
		// Если ничего не найдено, возвращаем ошибку, чтобы Usecase понял это
		return nil, err
	}
	return ad, nil
}

func (r *adRepository) Update(ctx context.Context, ad *domain.Ad) error {
	query := `
		UPDATE advertisements 
		SET title = $1, description = $2, price = $3, category_id = $4, location_id = NULLIF($5, 0), extra_props = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`
	_, err := r.storage.Pool.Exec(ctx, query,
		ad.Title, ad.Description, ad.Price, ad.CategoryID, ad.LocationID, ad.ExtraProps, ad.ID,
	)
	return err
}

func (r *adRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM advertisements WHERE id = $1`
	_, err := r.storage.Pool.Exec(ctx, query, id)
	return err
}

// ИЗБРАННОЕ
func (r *adRepository) AddToFavorites(ctx context.Context, userID, adID uuid.UUID) error {
	query := `INSERT INTO favorites (user_id, ad_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.storage.Pool.Exec(ctx, query, userID, adID)
	return err
}

func (r *adRepository) RemoveFromFavorites(ctx context.Context, userID, adID uuid.UUID) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND ad_id = $2`
	_, err := r.storage.Pool.Exec(ctx, query, userID, adID)
	return err
}

func (r *adRepository) GetFavorites(ctx context.Context, userID uuid.UUID) ([]*domain.Ad, error) {
	query := `
		SELECT a.id, a.user_id, a.title, a.price, a.deal_type, a.created_at 
		FROM advertisements a
		JOIN favorites f ON a.id = f.ad_id
		WHERE f.user_id = $1
	`
	rows, err := r.storage.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ads := make([]*domain.Ad, 0)
	for rows.Next() {
		ad := &domain.Ad{}
		rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Price, &ad.DealType, &ad.CreatedAt)
		ads = append(ads, ad)
	}
	return ads, nil
}
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

func (r *adRepository) UpdateStatus(ctx context.Context, adID uuid.UUID, status string, reason string) error {
	query := `UPDATE advertisements SET status = $1, rejection_reason = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.storage.Pool.Exec(ctx, query, status, reason, adID)
	return err
}

func (r *adRepository) CreateReport(ctx context.Context, rep *domain.Report) error {
	query := `INSERT INTO reports (id, reporter_id, ad_id, reason) VALUES ($1, $2, $3, $4)`

	if rep.ID == uuid.Nil {
		rep.ID = uuid.New()
	}

	_, err := r.storage.Pool.Exec(ctx, query, rep.ID, rep.ReporterID, rep.AdID, rep.Reason)
	return err
}

func (r *adRepository) GetReports(ctx context.Context) ([]*domain.Report, error) {
	query := `SELECT id, reporter_id, ad_id, reason, status FROM reports ORDER BY created_at DESC`
	rows, err := r.storage.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]*domain.Report, 0)
	for rows.Next() {
		rep := &domain.Report{}
		err := rows.Scan(&rep.ID, &rep.ReporterID, &rep.AdID, &rep.Reason, &rep.Status)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, nil
}
