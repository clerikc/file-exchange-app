package handlers

import (
	"file-exchange-app/storage"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

// FileInfo представляет информацию о файле
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
}

// DashboardHandler отображает главную страницу пользователя
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, _ := session.Values["username"].(string)
	canUpload, _ := session.Values["canUpload"].(bool)
	isAdmin, _ := session.Values["isAdmin"].(bool)

	// Получаем список файлов
	files, err := getFileList("./uploads")
	if err != nil {
		http.Error(w, "Error reading files", http.StatusInternalServerError)
		return
	}

	type TemplateData struct {
		Username  string
		CanUpload bool
		IsAdmin   bool
		Files     []FileInfo
	}

	data := TemplateData{
		Username:  username,
		CanUpload: canUpload,
		IsAdmin:   isAdmin,
		Files:     files,
	}

	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	tmpl.Execute(w, data)
}

// Вспомогательная функция для получения списка файлов
func getFileList(dir string) ([]FileInfo, error) {
	var files []FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Если папка не существует, создаем ее
			os.MkdirAll(dir, 0755)
			return files, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Name:    entry.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
	}
	return files, nil
}

// UploadHandler обрабатывает загрузку файлов
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, _ := session.Values["username"].(string)
	canUpload, _ := session.Values["canUpload"].(bool)

	if !canUpload {
		http.Error(w, "You don't have permission to upload files", http.StatusForbidden)
		return
	}

	// Ограничиваем размер файла (например, 100MB)
	r.ParseMultipartForm(100 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаем папку uploads если ее нет
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.MkdirAll("./uploads", 0755)
	}

	// Создаем файл на диске
	dst, err := os.Create(filepath.Join("./uploads", handler.Filename))
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Копируем содержимое файла
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Логируем действие
	_, err = storage.DB.Exec(
		"INSERT INTO logs (username, action, filename) VALUES (?, ?, ?)",
		username, "upload", handler.Filename,
	)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// DownloadHandler обрабатывает скачивание файлов
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, _ := session.Values["username"].(string)
	canDownload, _ := session.Values["canDownload"].(bool)

	if !canDownload {
		http.Error(w, "You don't have permission to download files", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	filename := vars["filename"]

	// Проверяем существование файла
	filePath := filepath.Join("./uploads", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	_, err := storage.DB.Exec(
		"INSERT INTO logs (username, action, filename) VALUES (?, ?, ?)",
		username, "download", filename,
	)
	if err != nil {
		log.Printf("Failed to log download action: %v", err)
		// Можно также вернуть ошибку или обработать её другим способом
	}

	// Отдаем файл пользователю
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}
