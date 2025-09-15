package models

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"` // Мы будем хранить хэш пароля, а не сам пароль!
	CanUpload    bool   `json:"can_upload"`
	CanDownload  bool   `json:"can_download"`
	IsAdmin      bool   `json:"is_admin"`
}

// Роли для упрощения (можно использовать вместо флагов)
const (
	RoleDownloader = "downloader" // Может только скачивать
	RoleUploader   = "uploader"   // Может и скачивать, и загружать
	RoleAdmin      = "admin"      // Полные права + админка
)
