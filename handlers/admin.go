package handlers

import (
	"file-exchange-app/storage"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// AdminHandler отображает админскую панель
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/admin.html"))

	// Получаем список всех пользователей из БД для отображения
	rows, err := storage.DB.Query("SELECT id, username, can_upload, can_download, is_admin FROM users")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type UserView struct {
		ID          int
		Username    string
		CanUpload   bool
		CanDownload bool
		IsAdmin     bool
		Role        string
	}

	var users []UserView
	for rows.Next() {
		var u UserView
		err := rows.Scan(&u.ID, &u.Username, &u.CanUpload, &u.CanDownload, &u.IsAdmin)
		if err != nil {
			continue
		}

		// Определяем роль для отображения
		if u.IsAdmin {
			u.Role = "Admin"
		} else if u.CanUpload && u.CanDownload {
			u.Role = "Uploader/Downloader"
		} else if u.CanDownload {
			u.Role = "Downloader"
		} else {
			u.Role = "No access"
		}
		users = append(users, u)
	}

	// Получаем логи для отображения
	logRows, err := storage.DB.Query("SELECT username, action, filename, timestamp FROM logs ORDER BY timestamp DESC LIMIT 100")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer logRows.Close()

	type LogEntry struct {
		Username  string
		Action    string
		Filename  string
		Timestamp string
	}

	var logs []LogEntry
	for logRows.Next() {
		var l LogEntry
		err := logRows.Scan(&l.Username, &l.Action, &l.Filename, &l.Timestamp)
		if err != nil {
			continue
		}
		logs = append(logs, l)
	}

	data := struct {
		Users []UserView
		Logs  []LogEntry
	}{
		Users: users,
		Logs:  logs,
	}

	tmpl.Execute(w, data)
}

// CreateUserHandler обрабатывает создание нового пользователя
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Валидация
	if username == "" || password == "" || role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Определяем права в зависимости от роли
	var canUpload, canDownload, isAdmin bool
	switch role {
	case "admin":
		canUpload, canDownload, isAdmin = true, true, true
	case "uploader":
		canUpload, canDownload, isAdmin = true, true, false
	case "downloader":
		canUpload, canDownload, isAdmin = false, true, false
	default:
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Вставляем пользователя в БД
	_, err = storage.DB.Exec(
		"INSERT INTO users (username, password_hash, can_upload, can_download, is_admin) VALUES (?, ?, ?, ?, ?)",
		username, string(hashedPassword), canUpload, canDownload, isAdmin,
	)

	if err != nil {
		// Если пользователь уже существует
		if err.Error() == "UNIQUE constraint failed: users.username" {
			http.Error(w, "Username already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Логируем действие
	session, _ := store.Get(r, "session-name")
	adminUser, _ := session.Values["username"].(string)
	_, err = storage.DB.Exec(
		"INSERT INTO logs (username, action, filename) VALUES (?, ?, ?)",
		adminUser, "create_user", fmt.Sprintf("Created user: %s with role: %s", username, role),
	)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
