package db

import (
    "context"
    "database/sql"
    "fmt"

    "library/internal/models"
)

type Database struct {
    conn *sql.DB
}

// New создает новый экземпляр Database
func New(conn *sql.DB) *Database {
    return &Database{conn: conn}
}

// GetBooks возвращает список книг с фильтрацией и пагинацией
func (db *Database) GetBooks(ctx context.Context, query *models.BooksQuery) ([]models.Book, error) {
    q := `SELECT id, title, author, isbn, year, status FROM books WHERE 1=1`
    args := []interface{}{}

    if query.Author != "" {
        q += " AND author LIKE ?"
        args = append(args, "%"+query.Author+"%")
    }
    if query.Status != "" {
        q += " AND status = ?"
        args = append(args, query.Status)
    }

    limit := query.Limit
    if limit <= 0 {
        limit = 10
    }
    page := query.Page
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * limit
    q += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

    rows, err := db.conn.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var books []models.Book
    for rows.Next() {
        var book models.Book
        if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Status); err != nil {
            return nil, err
        }
        books = append(books, book)
    }
    return books, rows.Err()
}

// GetBook возвращает книгу по ID
func (db *Database) GetBook(ctx context.Context, id string) (*models.Book, error) {
    var book models.Book
    err := db.conn.QueryRowContext(ctx, "SELECT id, title, author, isbn, year, status FROM books WHERE id = ?", id).
        Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Status)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &book, err
}

// CreateBook сохраняет новую книгу в базе данных
func (db *Database) CreateBook(ctx context.Context, book *models.Book) error {
    _, err := db.conn.ExecContext(ctx, "INSERT INTO books VALUES (?, ?, ?, ?, ?, ?)",
        book.ID, book.Title, book.Author, book.ISBN, book.Year, book.Status)
    return err
}

// UpdateBook обновляет книгу по ID
func (db *Database) UpdateBook(ctx context.Context, book *models.Book) error {
    _, err := db.conn.ExecContext(ctx, "UPDATE books SET title=?, author=?, isbn=?, year=?, status=? WHERE id=?",
        book.Title, book.Author, book.ISBN, book.Year, book.Status, book.ID)
    return err
}

// GetUser возвращает пользователя по ID
func (db *Database) GetUser(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    err := db.conn.QueryRowContext(ctx, "SELECT id, name, email, registration_date FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email, &user.RegistrationDate)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &user, err
}

// CreateUser сохраняет нового пользователя в базе данных
func (db *Database) CreateUser(ctx context.Context, user *models.User) error {
    _, err := db.conn.ExecContext(ctx, "INSERT INTO users VALUES (?, ?, ?, ?)",
        user.ID, user.Name, user.Email, user.RegistrationDate)
    return err
}

// IssueBook сохраняет запись о выдаче книги пользователю
func (db *Database) IssueBook(ctx context.Context, issue *models.Issue) error {
    _, err := db.conn.ExecContext(ctx, "INSERT INTO issues VALUES (?, ?, ?, ?, ?)",
        issue.BookID, issue.UserID, issue.IssueDate, issue.DueDate, issue.ReturnDate)
    return err
}

// ReturnBook обновляет дату возврата книги
func (db *Database) ReturnBook(ctx context.Context, bookID, returnDate string) error {
    _, err := db.conn.ExecContext(ctx, "UPDATE issues SET return_date=? WHERE book_id=? AND return_date IS NULL",
        returnDate, bookID)
    return err
}

// GetUserIssuedBooks возвращает книги, выданные пользователю
func (db *Database) GetUserIssuedBooks(ctx context.Context, userID string) ([]models.Book, error) {
    rows, err := db.conn.QueryContext(ctx, `
        SELECT b.id, b.title, b.author, b.isbn, b.year, b.status
        FROM books b
        INNER JOIN issues i ON b.id = i.book_id
        WHERE i.user_id = ? AND i.return_date IS NULL`, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var books []models.Book
    for rows.Next() {
        var book models.Book
        if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Status); err != nil {
            return nil, err
        }
        books = append(books, book)
    }
    return books, rows.Err()
}
