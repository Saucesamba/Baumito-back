package postgres

import (
	"Avito-back/internal/domain"
	"context"
	"github.com/google/uuid"
)

type userRepository struct {
	storage *Storage
}

// NewUserRepository создает новый экземпляр репозитория
func NewUserRepository(storage *Storage) domain.UserRepository {
	return &userRepository{storage: storage}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// Используем NULLIF($6, 0).
	// Это значит: если пришел 0, запиши в базу NULL.
	// NULL не проверяется по Foreign Key, и ошибка не вылетит.
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, university_id, faculty_id)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, 0), NULLIF($7, 0))
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	_, err := r.storage.Pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Phone,
		user.UniversityID, // Здесь может быть 0, Postgres сам поймет
		user.FacultyID,    // И здесь может быть 0
	)
	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.storage.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Реализацию остальных методов (GetByID, Update) добавим по мере необходимости
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (r *userRepository) Update(ctx context.Context, user *domain.User) error { return nil }
