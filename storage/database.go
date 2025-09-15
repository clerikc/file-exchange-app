package storage

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB
var UserStoreInstance UserStore

func InitDB() error {
	var err error
	// Открываем соединение с БД. Файл `data.db` будет создан в корне проекта.
	DB, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}

	// Создаем таблицу пользователей, если ее нет
	createUserTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        can_upload BOOLEAN DEFAULT FALSE,
        can_download BOOLEAN DEFAULT TRUE,
        is_admin BOOLEAN DEFAULT FALSE
    );
    `
	_, err = DB.Exec(createUserTable)
	if err != nil {
		return err
	}

	// Создаем таблицу логов, если ее нет
	createLogTable := `
    CREATE TABLE IF NOT EXISTS logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        action TEXT NOT NULL,
        filename TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err = DB.Exec(createLogTable)
	if err != nil {
		return err
	}

	// Создаем администратора по умолчанию, если пользователей нет
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		_, err = DB.Exec(`INSERT INTO users (username, password_hash, can_upload, can_download, is_admin) 
                         VALUES (?, ?, ?, ?, ?)`,
			"admin", string(hashedPassword), true, true, true)
		if err != nil {
			return err
		}
		log.Println("Создан пользователь admin по умолчанию. СРОЧНО СМЕНИТЕ ПАРОЛЬ!")
	}

	// Инициализируем UserStore
	UserStoreInstance = NewUserStore(DB)

	return nil
}
