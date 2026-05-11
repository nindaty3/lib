package models

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Year   int    `json:"year"`
	Status string `json:"status"`
}

type User struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	RegistrationDate string `json:"registration_date"`
}

type Issue struct {
	BookID     string  `json:"book_id"`
	UserID     string  `json:"user_id"`
	IssueDate  string  `json:"issue_date"`
	DueDate    string  `json:"due_date"`
	ReturnDate *string `json:"return_date"`
}

type BooksQuery struct {
	Page   int
	Limit  int
	Author string
	Status string
}
