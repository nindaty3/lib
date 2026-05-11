Простое API для работы с книгами, пользователями и выдачами.

Запуск

1. Скопируйте файл `.env.example` в `.env`
2. При необходимости измените значения в `.env`
3. Выполните: go mod tidy
4. Запустите сервер:
go run ./cmd


Книги
- `POST /books` — добавить книгу
- `GET /books` — получить список книг
- `GET /books/{id}` — получить книгу по ID
- `PUT /books/{id}` — обновить книгу

Пользователи
- `POST /users` — зарегистрировать пользователя
- `GET /users/{id}/books` — список книг, выданных пользователю

Выдача и возврат
- `POST /issues` — выдать книгу пользователю
- `POST /returns` — вернуть книгу

Примеры

Добавить книгу:

curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Book","author":"Author","isbn":"123","year":2023}'

Список книг:
curl http://localhost:8080/books

Зарегистрировать пользователя:
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"User","email":"user@example.com"}'

Получить книги пользователя:
curl http://localhost:8080/users/<user-id>/books

Выдать книгу:
curl -X POST http://localhost:8080/issues \
  -H "Content-Type: application/json" \
  -d '{"book_id":"book-id","user_id":"user-id"}'

Вернуть книгу:
curl -X POST http://localhost:8080/returns \
  -H "Content-Type: application/json" \
  -d '{"book_id":"book-id","user_id":"user-id"}'
