package handlers

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"

    "library/internal/db"
    "library/internal/models"
)

type Handlers struct {
    db *db.Database
}

// New создает новый набор HTTP обработчиков
func New(database *db.Database) *Handlers {
    return &Handlers{db: database}
}

func (h *Handlers) logger(r *http.Request) *slog.Logger {
    return slog.With("method", r.Method, "path", r.URL.Path)
}

// GetBooks возвращает список книг с пагинацией и фильтрацией
func (h *Handlers) GetBooks(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    query := &models.BooksQuery{
        Page:   parseInt(r.URL.Query().Get("page"), 1),
        Limit:  parseInt(r.URL.Query().Get("limit"), 10),
        Author: r.URL.Query().Get("author"),
        Status: r.URL.Query().Get("status"),
    }

    books, err := h.db.GetBooks(r.Context(), query)
    if err != nil {
        logger.Error("failed to list books", "error", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(books)
}

// CreateBook создает новую книгу из JSON запроса
func (h *Handlers) CreateBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    var book models.Book
    if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
        logger.Error("invalid request body", "error", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    book.ID = uuid.New().String()
    if book.Status == "" {
        book.Status = "Available"
    }

    if err := h.db.CreateBook(r.Context(), &book); err != nil {
        logger.Error("failed to create book", "error", err, "book_id", book.ID)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    logger.Info("created book", "book_id", book.ID)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(book)
}

// GetBook возвращает книгу по её ID
func (h *Handlers) GetBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    id := strings.TrimPrefix(r.URL.Path, "/books/")
    if id == "" {
        logger.Warn("missing book id")
        http.Error(w, "Book ID required", http.StatusBadRequest)
        return
    }

    book, err := h.db.GetBook(r.Context(), id)
    if err != nil {
        logger.Error("failed to get book", "error", err, "book_id", id)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if book == nil {
        logger.Warn("book not found", "book_id", id)
        http.Error(w, "Book not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(book)
}

// UpdateBook обновляет книгу по её ID
func (h *Handlers) UpdateBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    id := strings.TrimPrefix(r.URL.Path, "/books/")
    if id == "" {
        logger.Warn("missing book id")
        http.Error(w, "Book ID required", http.StatusBadRequest)
        return
    }

    var book models.Book
    if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
        logger.Error("invalid request body", "error", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    book.ID = id
    if err := h.db.UpdateBook(r.Context(), &book); err != nil {
        logger.Error("failed to update book", "error", err, "book_id", id)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    logger.Info("updated book", "book_id", id)
    json.NewEncoder(w).Encode(book)
}

// RegisterUser создаёт нового пользователя из JSON запроса
func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        logger.Error("invalid request body", "error", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    user.ID = uuid.New().String()
    user.RegistrationDate = time.Now().Format("2006-01-02")

    if err := h.db.CreateUser(r.Context(), &user); err != nil {
        logger.Error("failed to register user", "error", err, "user_id", user.ID)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    logger.Info("registered user", "user_id", user.ID)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

// GetUserBooks возвращает пользователя и его выданные книги
func (h *Handlers) GetUserBooks(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    trimmed := strings.TrimPrefix(r.URL.Path, "/users/")
    if !strings.HasSuffix(trimmed, "/books") {
        logger.Warn("invalid user books path", "path", r.URL.Path)
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    id := strings.TrimSuffix(trimmed, "/books")
    if id == "" {
        logger.Warn("missing user id")
        http.Error(w, "User ID required", http.StatusBadRequest)
        return
    }

    user, err := h.db.GetUser(r.Context(), id)
    if err != nil {
        logger.Error("failed to get user", "error", err, "user_id", id)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if user == nil {
        logger.Warn("user not found", "user_id", id)
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    books, err := h.db.GetUserIssuedBooks(r.Context(), id)
    if err != nil {
        logger.Error("failed to list user books", "error", err, "user_id", id)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{"user": user, "books": books})
}

// IssueBook создает запись о выдаче книги пользователю
func (h *Handlers) IssueBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    var issue models.Issue
    if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
        logger.Error("invalid request body", "error", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    issue.IssueDate = time.Now().Format("2006-01-02")
    if issue.DueDate == "" {
        issue.DueDate = time.Now().AddDate(0, 0, 14).Format("2006-01-02")
    }

    if err := h.db.IssueBook(r.Context(), &issue); err != nil {
        logger.Error("failed to issue book", "error", err, "book_id", issue.BookID, "user_id", issue.UserID)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    book, _ := h.db.GetBook(r.Context(), issue.BookID)
    if book != nil {
        book.Status = "Issued"
        h.db.UpdateBook(r.Context(), book)
    }

    logger.Info("book issued", "book_id", issue.BookID, "user_id", issue.UserID)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(issue)
}

// ReturnBook обрабатывает возврат книги и обновляет её статус
func (h *Handlers) ReturnBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger(r)
    w.Header().Set("Content-Type", "application/json")

    var req map[string]string
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error("invalid request body", "error", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    bookID, ok := req["book_id"]
    if !ok {
        logger.Warn("book_id missing")
        http.Error(w, "book_id required", http.StatusBadRequest)
        return
    }

    returnDate := time.Now().Format("2006-01-02")
    if err := h.db.ReturnBook(r.Context(), bookID, returnDate); err != nil {
        logger.Error("failed to return book", "error", err, "book_id", bookID)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    book, _ := h.db.GetBook(r.Context(), bookID)
    if book != nil {
        book.Status = "Available"
        h.db.UpdateBook(r.Context(), book)
    }

    logger.Info("book returned", "book_id", bookID)
    json.NewEncoder(w).Encode(map[string]string{"status": "Book returned"})
}

// parseInt парсит строку в число и возвращает значение по умолчанию при ошибке
func parseInt(s string, defaultValue int) int {
    if s == "" {
        return defaultValue
    }
    if val, err := strconv.Atoi(s); err == nil && val > 0 {
        return val
    }
    return defaultValue
}
