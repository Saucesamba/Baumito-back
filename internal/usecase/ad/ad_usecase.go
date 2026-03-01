package ad

import (
	"Avito-back/internal/domain"
	"Avito-back/internal/repository/redis"
	"Avito-back/internal/repository/s3"
	"context"
	"io"

	"github.com/google/uuid"
)

type adUsecase struct {
	adRepo    domain.AdRepository
	s3Repo    *s3.FileRepository
	cacheRepo *redis.AdCacheRepository // Добавь это поле
}

func NewAdUsecase(repo domain.AdRepository, s3 *s3.FileRepository, cache *redis.AdCacheRepository) domain.AdUsecase {
	return &adUsecase{
		adRepo:    repo,
		s3Repo:    s3,
		cacheRepo: cache,
	}
}

func (u *adUsecase) CreateAd(ctx context.Context, ad *domain.Ad) error {
	// Здесь можно добавить проверку: если цена 0, а тип сделки "sell" — выдать ошибку
	return u.adRepo.Create(ctx, ad)
}

func (u *adUsecase) GetAd(ctx context.Context, id uuid.UUID) (*domain.Ad, error) {
	// 1. Увеличиваем счетчик в Redis (Highload подход)
	_ = u.cacheRepo.IncrementViews(ctx, id.String())

	// 2. Получаем данные из базы
	ad, err := u.adRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Подменяем количество просмотров на актуальное из кэша
	views, _ := u.cacheRepo.GetViews(ctx, id.String())
	ad.ViewsCount = int(views)

	return ad, nil
}

func (u *adUsecase) ListAds(ctx context.Context, filter domain.AdFilter, page int) ([]*domain.Ad, error) {
	if page <= 0 {
		page = 1
	}
	filter.Limit = 20
	filter.Offset = (page - 1) * filter.Limit

	return u.adRepo.Fetch(ctx, filter)
}

func (u *adUsecase) UploadImage(ctx context.Context, adID uuid.UUID, fileName string, file io.Reader, size int64) error {
	// 1. Грузим в MinIO
	path, err := u.s3Repo.Upload(ctx, fileName, file, size)
	if err != nil {
		return err
	}

	// 2. Сохраняем ссылку в Postgres
	img := &domain.AdImage{
		AdID:     adID,
		ImageURL: path,
		IsMain:   false, // Можно добавить логику проверки первой картинки
	}
	return u.adRepo.AddImage(ctx, img)
}
