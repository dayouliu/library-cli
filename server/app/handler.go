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
	response := api.Response{
		Type:       "error",
		StatusCode: statusCode,
		Message:    message + "\n" + err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
	}
	if data != nil {
		response.Data = data
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
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

	_, err = h.db.Exec(`INSERT INTO books (title, author, published_at, edition, description, genre) VALUES ($1, $2, $3, $4, $5, $6)`,
		book.Title, book.Author, book.PublishedAt.Format("2006-01-02"), book.Edition, book.Description, book.Genre)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error creating book")
		return
	}

	respondJSON(w, nil)
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

	updateFields := make([]string, 0)
	if book.Author != "" {
		updateFields = append(updateFields, fmt.Sprintf("author = '%s'", book.Author))
	}
	if !book.PublishedAt.IsZero() {
		updateFields = append(updateFields, fmt.Sprintf("published_at = '%s'", book.PublishedAt.Format("2006-01-02")))
	}
	if book.Edition != "" {
		updateFields = append(updateFields, fmt.Sprintf("edition = '%s'", book.Edition))
	}
	if book.Description != "" {
		updateFields = append(updateFields, fmt.Sprintf("description = '%s'", book.Description))
	}
	if book.Genre != "" {
		updateFields = append(updateFields, fmt.Sprintf("genre = '%s'", book.Genre))
	}
	updateQuery := "UPDATE books SET " + strings.Join(updateFields, ", ") + " WHERE title = $1"

	_, err = h.db.Exec(updateQuery, book.Title)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error updating book")
		return
	}

	respondJSON(w, nil)
}

// removeBook removes a book from the database
func (h *Handler) removeBook(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respondError(w, nil, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	_, err := h.db.Exec(`DELETE FROM books WHERE title = $1`, title)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book")
		return
	}

	// delete book from collection_subscriptions
	_, err = h.db.Exec(`DELETE FROM collection_subscriptions WHERE book_title = $1`, title)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book from collection_subscriptions")
		return
	}

	respondJSON(w, nil)
}

// getBooks returns all books
func (h *Handler) getBooks(w http.ResponseWriter, r *http.Request) {
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
	query := "SELECT title, author, published_at, edition, description, genre FROM books"
	conditions := []string{}
	if title != "" {
		conditions = append(conditions, fmt.Sprintf("title = '%s'", title))
	}
	if genre != "" {
		conditions = append(conditions, fmt.Sprintf("genre = '%s'", genre))
	}
	if author != "" {
		conditions = append(conditions, fmt.Sprintf("author = '%s'", author))
	}
	if publishStartDate != "" {
		conditions = append(conditions, fmt.Sprintf("published_at >= '%s'", publishStartDate))
	}
	if publishEndDate != "" {
		conditions = append(conditions, fmt.Sprintf("published_at <= '%s'", publishEndDate))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := h.db.Query(query)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error getting books")
		return
	}
	defer rows.Close()

	books := make([]api.Book, 0)

	for rows.Next() {
		var book api.Book
		err := rows.Scan(&book.Title, &book.Author, &book.PublishedAt, &book.Edition, &book.Description, &book.Genre)
		if err != nil {
			respondError(w, err, http.StatusInternalServerError, "Error getting books")
			return
		}
		books = append(books, book)
	}

	respondJSON(w, books)
}

// createCollection creates a collection
func (h *Handler) createCollection(w http.ResponseWriter, r *http.Request) {
	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")
	fmt.Println(string(collectionName))

	_, err := h.db.Exec(`INSERT INTO collections (name) VALUES ($1)`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error creating collection")
		return
	}

	respondJSON(w, nil)
}

// removeCollection removes a collection
func (h *Handler) removeCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")
	if collectionName == "" {
		respondError(w, nil, http.StatusBadRequest, "collection_name cannot be empty")
		return
	}

	_, err := h.db.Exec(`DELETE FROM collections WHERE name = $1`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing collection")
		return
	}

	// remove all books in the collection subscription
	_, err = h.db.Exec(`DELETE FROM collection_subscriptions WHERE collection_name = $1`, collectionName)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing collection")
		return
	}

	respondJSON(w, nil)
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

	respondJSON(w, collections)
}

// addBookToCollection adds a book to a collection
func (h *Handler) addToCollection(w http.ResponseWriter, r *http.Request) {
	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := h.db.Exec(`INSERT INTO collection_subscriptions(collection_name, book_title) VALUES ($1, $2)`, collectionName, bookTitle)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error adding book to collection")
		return
	}

	respondJSON(w, nil)
}

// removeFromCollection removes a book from a collection
func (h *Handler) removeFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := h.db.Exec(`DELETE FROM collection_subscriptions WHERE collection_name = $1 AND book_title = $2`, collectionName, bookTitle)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError, "Error removing book from collection")
		return
	}

	respondJSON(w, nil)
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

	respondJSON(w, books)
}
