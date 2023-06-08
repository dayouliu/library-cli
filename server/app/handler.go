package app

import (
	"bms/shared/api"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Handler struct {
	db *sql.DB
}

func respondError(w http.ResponseWriter, err error, statusCode int, message string) {
	if err != nil {
		message = message + "\n" + err.Error()
	}

	response := api.Response{
		Type:       "error",
		StatusCode: statusCode,
		Message:    message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func respondJSON(w http.ResponseWriter, data interface{}, message string, statusCode int) {
	response := api.Response{
		Type:       "success",
		StatusCode: statusCode,
		Message:    message,
	}
	if data != nil {
		response.Data = data
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func genSQLConditions(conditions *[]string, values *[]any, op string, field string, value string, counter *int) {
	*conditions = append(*conditions, fmt.Sprintf("%s %s $%d", field, op, *counter))
	*values = append(*values, value)
	*counter++
}

func (h *Handler) createBook(w http.ResponseWriter, r *http.Request) {
	var book api.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		respondError(w, err, http.StatusBadRequest, "Invalid request body")
		return
	}

	if book.Title == "" {
		respondError(w, err, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	_, err = h.db.Exec(
		"INSERT INTO books (title, author, publish_date, edition, description, genre) VALUES ($1, $2, $3, $4, $5, $6)",
		book.Title, book.Author, book.PublishDate.Format("2006-01-02"), book.Edition, book.Description, book.Genre)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error creating book")
		return
	}

	respondJSON(w, nil, "Book created successfully", http.StatusCreated)
}

// addBook adds a book to the database
func (h *Handler) setBook(w http.ResponseWriter, r *http.Request) {
	var book api.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		respondError(w, err, http.StatusBadRequest, "Invalid request body")
		return
	}

	if book.Title == "" {
		respondError(w, err, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	counter := 1
	conditions := make([]string, 0)
	values := make([]any, 0)
	if book.Author != "" {
		genSQLConditions(&conditions, &values, "=", "author", book.Author, &counter)
	}
	if !book.PublishDate.IsZero() {
		genSQLConditions(&conditions, &values, "=", "publish_date", book.PublishDate.Format(api.PublishTimeLayoutDMY), &counter)
	}
	if book.Edition != "" {
		genSQLConditions(&conditions, &values, "=", "edition", book.Edition, &counter)
	}
	if book.Description != "" {
		genSQLConditions(&conditions, &values, "=", "description", book.Description, &counter)
	}
	if book.Genre != "" {
		genSQLConditions(&conditions, &values, "=", "genre", book.Genre, &counter)
	}

	if len(conditions) == 0 {
		respondError(w, err, http.StatusBadRequest, "No fields to update")
		return
	}

	updateQuery :=
		fmt.Sprintf("UPDATE books SET "+strings.Join(conditions, ", ")+" WHERE title = $%d", counter)
	values = append(values, book.Title)

	_, err = h.db.Exec(updateQuery, values...)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error updating book")
		return
	}

	respondJSON(w, nil, "Book updated successfully", http.StatusOK)
}

// removeBook removes a book from the database
func (h *Handler) removeBook(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respondError(w, nil, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	// delete book subscriptions from collection_subscriptions table first
	_, err := h.db.Exec(`DELETE FROM collection_subscriptions WHERE book_title = $1`, title)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book from collection_subscriptions")
		return
	}

	// remove book from books table
	_, err = h.db.Exec(`DELETE FROM books WHERE title = $1`, title)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book")
		return
	}

	respondJSON(w, nil, "Book removed successfully", http.StatusOK)
}

// listBooks returns all books
func (h *Handler) listBooks(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	genre := r.URL.Query().Get("genre")
	author := r.URL.Query().Get("author")
	publishStartDate := r.URL.Query().Get("publish_start")
	publishEndDate := r.URL.Query().Get("publish_end")

	if publishStartDate != "" && publishEndDate != "" && publishStartDate > publishEndDate {
		respondError(w, nil, http.StatusBadRequest, "publish_start cannot be greater than publish_end")
		return
	}

	// add filter conditions to query
	query := "SELECT title, author, publish_date, edition, description, genre FROM books"
	conditions := []string{}
	values := []any{}
	counter := 1
	if title != "" {
		genSQLConditions(&conditions, &values, "=", "title", title, &counter)
	}
	if genre != "" {
		genSQLConditions(&conditions, &values, "=", "genre", genre, &counter)
	}
	if author != "" {
		genSQLConditions(&conditions, &values, "=", "author", author, &counter)
	}
	if publishStartDate != "" {
		genSQLConditions(&conditions, &values, ">=", "publish_date", publishStartDate, &counter)
	}
	if publishEndDate != "" {
		genSQLConditions(&conditions, &values, "<=", "publish_date", publishEndDate, &counter)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := h.db.Query(query, values...)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error getting books")
		return
	}
	defer rows.Close()

	books := make([]api.Book, 0)

	for rows.Next() {
		var book api.Book
		err := rows.Scan(&book.Title, &book.Author, &book.PublishDate, &book.Edition, &book.Description, &book.Genre)
		if err != nil {
			respondError(w, err, http.StatusInternalServerError, "Error getting books")
			return
		}
		books = append(books, book)
	}

	respondJSON(w, books, "Books retrieved successfully", http.StatusOK)
}

// createCollection creates a collection
func (h *Handler) createCollection(w http.ResponseWriter, r *http.Request) {
	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")

	_, err := h.db.Exec(`INSERT INTO collections (name) VALUES ($1)`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error creating collection")
		return
	}

	respondJSON(w, nil, "Collection created successfully", http.StatusOK)
}

// removeCollection removes a collection
func (h *Handler) removeCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")

	if collectionName == "" {
		respondError(w, nil, http.StatusBadRequest, "collection_name cannot be empty")
		return
	}

	// remove all subscribed books in collection_subscription table first
	_, err := h.db.Exec(`DELETE FROM collection_subscriptions WHERE collection_name = $1`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing collection")
		return
	}

	// remove collection in collections table
	_, err = h.db.Exec(`DELETE FROM collections WHERE name = $1`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing collection")
		return
	}

	respondJSON(w, nil, "Collection removed successfully", http.StatusCreated)
}

// getCollections returns all books in a collection
func (h *Handler) getCollections(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT name FROM collections")
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error getting collections")
		return
	}
	defer rows.Close()

	collections := make([]string, 0)

	for rows.Next() {
		var collection string
		err := rows.Scan(&collection)
		if err != nil {
			respondError(w, err, http.StatusInternalServerError, "Error getting collections")
			return
		}
		collections = append(collections, collection)
	}

	respondJSON(w, collections, "Collections retrieved successfully", http.StatusOK)
}

// addBookToCollection adds a book to a collection
func (h *Handler) addBookToCollection(w http.ResponseWriter, r *http.Request) {
	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := h.db.Exec(`INSERT INTO collection_subscriptions(collection_name, book_title) VALUES ($1, $2)`, collectionName, bookTitle)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error adding book to collection")
		return
	}

	respondJSON(w, nil, "Book added to collection successfully", http.StatusOK)
}

// removeBookFromCollection removes a book from a collection
func (h *Handler) removeBookFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := h.db.Exec(`DELETE FROM collection_subscriptions WHERE collection_name = $1 AND book_title = $2`, collectionName, bookTitle)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book from collection")
		return
	}

	respondJSON(w, nil, "Book removed from collection successfully", http.StatusOK)
}

// getBooksInCollection returns all books in a collection
func (h *Handler) getBooksInCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")

	rows, err := h.db.Query("SELECT book_title FROM collection_subscriptions WHERE collection_name = $1", collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error getting books in collection")
		return
	}
	defer rows.Close()

	books := make([]string, 0)
	for rows.Next() {
		var book string
		err := rows.Scan(&book)
		if err != nil {
			respondError(w, err, http.StatusInternalServerError, "Error getting books in collection")
			return
		}
		books = append(books, book)
	}

	respondJSON(w, books, "Books in collection retrieved successfully", http.StatusOK)
}
