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
	// Добавляем role в список колонок и $8 в VALUES.
	// COALESCE(NULLIF($8, ''), 'user') — если роль пустая, ставим 'user' по умолчанию.
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, university_id, faculty_id, role)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, 0), NULLIF($7, 0), COALESCE(NULLIF($8, ''), 'user'))
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
		user.UniversityID,
		user.FacultyID,
		user.Role, // Передаем роль из структуры domain.User
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
	query := `SELECT id, email, name, role, is_blocked, university_id FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.storage.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role, &user.IsBlocked, &user.UniversityID,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error { return nil }

func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, isBlocked bool) error {
	query := `UPDATE users SET is_blocked = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.storage.Pool.Exec(ctx, query, isBlocked, userID)
	return err
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	// Выбираем основные поля пользователей для списка админа
	query := `
		SELECT id, email, name, role, is_blocked, university_id, created_at 
		FROM users 
		ORDER BY created_at DESC
	`

	rows, err := r.storage.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u := &domain.User{}
		err := rows.Scan(
			&u.ID, &u.Email, &u.Name, &u.Role,
			&u.IsBlocked, &u.UniversityID, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
