package main

import (
    "database/sql"
    "log/slog"
    "net/http"
    "os"

    _ "github.com/mattn/go-sqlite3"
    "github.com/joho/godotenv"

    "library/internal/db"
    "library/internal/handlers"
)

// main инициализирует приложение и запускает HTTP сервер
func main() {
    godotenv.Load()

    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
    slog.SetDefault(logger)

    dbPath := getEnv("DB_PATH", "library.db")
    port := getEnv("PORT", "8080")
    logger.Info("starting service", "port", port, "db_path", dbPath)

    conn, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        logger.Error("failed to open database", "error", err)
        os.Exit(1)
    }
    defer conn.Close()

    if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS books (
        id TEXT PRIMARY KEY, title TEXT NOT NULL, author TEXT NOT NULL,
        isbn TEXT NOT NULL, year INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'Available'
    )`); err != nil {
        logger.Error("failed to create books table", "error", err)
        os.Exit(1)
    }
    if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, registration_date TEXT NOT NULL
    )`); err != nil {
        logger.Error("failed to create users table", "error", err)
        os.Exit(1)
    }
    if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS issues (
        book_id TEXT NOT NULL, user_id TEXT NOT NULL, issue_date TEXT NOT NULL,
        due_date TEXT NOT NULL, return_date TEXT, FOREIGN KEY (book_id) REFERENCES books(id)
    )`); err != nil {
        logger.Error("failed to create issues table", "error", err)
        os.Exit(1)
    }

    database := db.New(conn)
    h := handlers.New(database)

    http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            h.GetBooks(w, r)
        case http.MethodPost:
            h.CreateBook(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            h.GetBook(w, r)
        case http.MethodPut:
            h.UpdateBook(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            h.RegisterUser(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            h.GetUserBooks(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            h.IssueBook(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/returns", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            h.ReturnBook(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    if err := http.ListenAndServe(":"+port, nil); err != nil {
        logger.Error("server stopped", "error", err)
        os.Exit(1)
    }
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
