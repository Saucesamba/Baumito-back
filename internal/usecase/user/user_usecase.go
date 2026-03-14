package user

import (
	"Avito-back/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: repo}
}

func (u *userUsecase) Register(ctx context.Context, user *domain.User, password string) error {
	// 1. Хешируем пароль (Highload стандарт безопасности)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// 2. Сохраняем в базу через репозиторий
	return u.userRepo.Create(ctx, user)
}

const jwtSecret = "your-super-secret-key-for-campus" //TODO В идеале тянуть из config

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Ищем пользователя по email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 2. Сравниваем хеш пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	// 3. Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),                          // ID пользователя
		"role": user.Role,                                 // ТЕПЕРЬ РОЛЬ ВНУТРИ JWT
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // Токен живет неделю
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *userUsecase) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (u *userUsecase) VerifyStudent(ctx context.Context, id uuid.UUID) error { return nil }

func (u *userUsecase) BlockUser(ctx context.Context, userID uuid.UUID) error {
	return u.userRepo.UpdateStatus(ctx, userID, true) // ставим true в is_blocked
}
