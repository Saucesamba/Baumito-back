package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User - модель студента в системе
type User struct {
	ID                uuid.UUID `json:"id"`
	UniversityID      int       `json:"university_id"`
	FacultyID         int       `json:"faculty_id"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone,omitempty"`
	PasswordHash      string    `json:"-"` // Пароль скрываем от JSON
	Name              string    `json:"name"`
	AvatarURL         string    `json:"avatar_url,omitempty"`
	StudentCardID     string    `json:"student_card_id,omitempty"`
	IsVerifiedStudent bool      `json:"is_verified_student"`
	Rating            float64   `json:"rating"`
	IsBlocked         bool      `json:"is_blocked"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserRepository - интерфейс для работы с БД (слой Repository)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

// UserUsecase - интерфейс бизнес-логики (слой Usecase)
type UserUsecase interface {
	Register(ctx context.Context, user *User, password string) error
	Login(ctx context.Context, email, password string) (string, error) // Возвращает JWT токен
	GetProfile(ctx context.Context, id uuid.UUID) (*User, error)
	VerifyStudent(ctx context.Context, id uuid.UUID) error
}
