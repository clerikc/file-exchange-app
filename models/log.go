package models

import "time"

// LogEntry представляет запись в логе действий
type LogEntry struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`   // 'login_success', 'login_failed', 'upload', 'download', 'create_user'
	Filename  string    `json:"filename"` // имя файла или описание действия
	Timestamp time.Time `json:"timestamp"`
}

// LogAction типы действий для логирования
const (
	ActionLoginSuccess = "login_success"
	ActionLoginFailed  = "login_failed"
	ActionUpload       = "upload"
	ActionDownload     = "download"
	ActionCreateUser   = "create_user"
	ActionDeleteFile   = "delete_file"
)
