package storage

import (
	"database/sql"
	"file-exchange-app/models"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// UserStore представляет интерфейс для работы с пользователями
type UserStore interface {
	CreateUser(username, password string, canUpload, canDownload, isAdmin bool) error
	GetUserByUsername(username string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	VerifyUserCredentials(username, password string) (*models.User, error)
	DeleteUser(userID int) error
}

// SQLiteUserStore реализация UserStore для SQLite
type SQLiteUserStore struct {
	db *sql.DB
}

// NewUserStore создает новый экземпляр UserStore
func NewUserStore(db *sql.DB) UserStore {
	return &SQLiteUserStore{db: db}
}

// CreateUser создает нового пользователя в базе данных
func (s *SQLiteUserStore) CreateUser(username, password string, canUpload, canDownload, isAdmin bool) error {
	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Вставляем пользователя в БД
	_, err = s.db.Exec(
		"INSERT INTO users (username, password_hash, can_upload, can_download, is_admin) VALUES (?, ?, ?, ?, ?)",
		username, string(hashedPassword), canUpload, canDownload, isAdmin,
	)

	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return fmt.Errorf("username already exists")
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

// GetUserByUsername возвращает пользователя по имени
func (s *SQLiteUserStore) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, can_upload, can_download, is_admin FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CanUpload, &user.CanDownload, &user.IsAdmin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// GetAllUsers возвращает всех пользователей
func (s *SQLiteUserStore) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query("SELECT id, username, can_upload, can_download, is_admin FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.CanUpload, &user.CanDownload, &user.IsAdmin)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// VerifyUserCredentials проверяет логин и пароль пользователя
func (s *SQLiteUserStore) VerifyUserCredentials(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err // Пользователь не найден
	}

	// Сравниваем пароль с хэшем
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// DeleteUser удаляет пользователя по ID
func (s *SQLiteUserStore) DeleteUser(userID int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}
