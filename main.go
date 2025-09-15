package main

import (
	"file-exchange-app/handlers"
	"file-exchange-app/storage"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Объявляем метрики как глобальные переменные
var (
	// Счетчик попыток авторизации с меткой status (success/failure)
	loginAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "file_exchange_login_attempts_total",
		Help: "Total number of login attempts",
	}, []string{"status"})

	// Счетчик операций с файлами с метками type (upload/download) и status (success/failure)
	fileOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "file_exchange_file_operations_total",
		Help: "Total number of file operations",
	}, []string{"type", "status"})

	// Gauge для отслеживания использования дискового пространства
	diskUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "file_exchange_disk_usage_bytes",
		Help: "Current disk usage of uploads directory in bytes",
	})

	// Счетчик для отслеживания количества файлов
	fileCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "file_exchange_file_count",
		Help: "Current number of files in uploads directory",
	})
)

// Функция для расчета размера директории
func getDirSize(path string) (int64, int, error) {
	var size int64
	var count int
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})
	return size, count, err
}

// Функция для периодического обновления метрик диска
func updateDiskMetrics() {
	for {
		size, count, err := getDirSize("./uploads")
		if err != nil {
			if os.IsNotExist(err) {
				// Папка не существует, устанавливаем 0
				diskUsageBytes.Set(0)
				fileCount.Set(0)
			} else {
				log.Printf("Error getting directory size: %v", err)
			}
		} else {
			diskUsageBytes.Set(float64(size))
			fileCount.Set(float64(count))
		}
		time.Sleep(30 * time.Second) // Обновляем каждые 30 секунд
	}
}

func main() {
	// Инициализируем БД
	err := storage.InitDB()
	if err != nil {
		log.Fatal("Could not initialize database:", err)
	}

	// Запускаем горутину для обновления метрик диска
	go updateDiskMetrics()

	r := mux.NewRouter()

	// Публичные маршруты
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Защищенные маршруты (требуют авторизации) - ИСПРАВЛЕНО
	r.Handle("/dashboard", handlers.AuthMiddleware(http.HandlerFunc(handlers.DashboardHandler))).Methods("GET")
	r.Handle("/upload", handlers.AuthMiddleware(http.HandlerFunc(handlers.UploadHandler))).Methods("POST")
	r.Handle("/download/{filename}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DownloadHandler))).Methods("GET")

	// Админские маршруты (требуют прав администратора)
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(handlers.AdminMiddleware)
	adminRouter.HandleFunc("", handlers.AdminHandler).Methods("GET")
	adminRouter.HandleFunc("/create-user", handlers.CreateUserHandler).Methods("POST")

	// Маршрут для метрик Prometheus
	r.Handle("/metrics", promhttp.Handler())

	// Health check endpoint для мониторинга
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Устанавливаем метрики в handlers
	//handlers.LoginAttemptsCounter = loginAttempts
	//handlers.FileOperationsCounter = fileOperations

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
