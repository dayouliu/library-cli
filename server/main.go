package main

import (
	"bms/shared/api"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	_ "strconv"
	"strings"
)

var db *sql.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "bms_db"
)

func main() {
	// Open a database connection
	var err error
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the books and collections tables if they don't exist
	createTables()

	// Initialize the router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// add book
	router.Post("/book/create", createBook)

	router.Get("/book/list", getBooks)

	// set book
	router.Post("/book/set", setBook)

	// collections
	router.Post("/collection/create", createCollection)
	router.Post("/collection/add", addToCollection)
	router.Post("/collection/remove", removeFromCollection)
	router.Get("/collection/list", getCollections)
	router.Post("/collection/list/books", getBooksInCollection)

	// Start the server
	server := http.Server{
		Addr:    ":8080", // Specify the address where your server is running
		Handler: router,
	}

	server.ListenAndServe()

	//// listen to stop signals
	//quit := make(chan os.Signal, 1)
	//signal.Notify(quit, os.Interrupt)
	//<-quit
	//log.Println("Shutting down server...")
	//
	//log.Println("Server gracefully exited")
}

// createTables creates the books and collections tables if they don't exist
func createTables() {

	// unique non empty string title
	createBooksTableQuery := `CREATE TABLE IF NOT EXISTS books (
		title VARCHAR(255) NOT NULL PRIMARY KEY,
		author VARCHAR(255),
		published_at DATE,
		edition VARCHAR(10),
		description TEXT,
		genre VARCHAR(255)
	);`

	createCollectionsTableQuery := `CREATE TABLE IF NOT EXISTS collections (
		name VARCHAR(255) NOT NULL PRIMARY KEY,
    	description TEXT
	);`

	createCollectionSubscriptions := `CREATE TABLE IF NOT EXISTS collection_subscriptions (
    	book_title VARCHAR(255),
    	collection_name VARCHAR(255),
    	PRIMARY KEY (book_title, collection_name),
		FOREIGN KEY (book_title) REFERENCES books (title), 
    	FOREIGN KEY (collection_name) REFERENCES collections (name)
	);`

	_, err := db.Exec(createBooksTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCollectionsTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCollectionSubscriptions)
	if err != nil {
		log.Fatal(err)
	}
}

// addBook adds a book to the database
func createBook(w http.ResponseWriter, r *http.Request) {
	var book api.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		handleError(w, err, http.StatusBadRequest, "Invalid request body")
		return
	}

	if book.Title == "" {
		handleError(w, err, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	_, err = db.Exec(`INSERT INTO books (title, author, published_at, edition, description, genre) VALUES ($1, $2, $3, $4, $5, $6)`,
		book.Title, book.Author, book.PublishedAt.Format("2006-01-02"), book.Edition, book.Description, book.Genre)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error creating book")
		return
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
		Data:       nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// addBook adds a book to the database
func setBook(w http.ResponseWriter, r *http.Request) {
	var book api.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		handleError(w, err, http.StatusBadRequest, "Invalid request body")
		return
	}

	if book.Title == "" {
		handleError(w, err, http.StatusBadRequest, "Title cannot be empty")
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

	_, err = db.Exec(updateQuery, book.Title)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error updating book")
		return
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
		Data:       nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getBooks returns all books
func getBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT title, author, published_at, edition, description, genre FROM books")
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error getting books")
		return
	}
	defer rows.Close()

	books := make([]api.Book, 0)

	for rows.Next() {
		var book api.Book
		err := rows.Scan(&book.Title, &book.Author, &book.PublishedAt, &book.Edition, &book.Description, &book.Genre)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError, "Error getting books")
			return
		}
		books = append(books, book)
	}

	// data is a map that contains key-value pairs
	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusOK,
		Data:       books,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// createCollection creates a collection
func createCollection(w http.ResponseWriter, r *http.Request) {
	// print all url params
	fmt.Println(r.URL.Query())

	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")
	fmt.Println(string(collectionName))

	_, err := db.Exec(`INSERT INTO collections (name) VALUES ($1)`, collectionName)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error creating collection")
		return
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
		Data:       nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getCollections returns all books in a collection
func getCollections(w http.ResponseWriter, r *http.Request) {
	fmt.Println("!!!")

	rows, err := db.Query("SELECT name FROM collections")
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error getting collections")
		return
	}
	defer rows.Close()

	collections := make([]api.Collection, 0)

	for rows.Next() {
		var collection api.Collection
		err := rows.Scan(&collection.Name)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError, "Error getting collections")
			return
		}
		collections = append(collections, collection)
	}

	// data is a map that contains key-value pairs
	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusOK,
		Data:       collections,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// addBookToCollection adds a book to a collection
func addToCollection(w http.ResponseWriter, r *http.Request) {
	// get parameter from URL with chi library
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := db.Exec(`INSERT INTO collection_subscriptions(collection_name, book_title) VALUES ($1, $2)`, collectionName, bookTitle)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error adding book to collection")
		return
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
		Data:       nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// removeBookFromCollection removes a book from a collection
func removeFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")
	bookTitle := r.URL.Query().Get("book_title")

	_, err := db.Exec(`DELETE FROM collection_subscriptionsWHERE collection_name = $1 AND book_title = $2`, collectionName, bookTitle)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error removing book from collection")
		return
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusCreated,
		Data:       nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getBooksInCollection returns all books in a collection
func getBooksInCollection(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Query().Get("collection_name")

	rows, err := db.Query("SELECT book_title FROM collection_subscriptions WHERE collection_name = $1", collectionName)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError, "Error getting books in collection")
		return
	}
	defer rows.Close()

	books := make([]string, 0)
	for rows.Next() {
		var book string
		err := rows.Scan(&book)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError, "Error getting books in collection")
			return
		}
		books = append(books, book)
	}

	response := api.Response{
		Type:       "success",
		StatusCode: http.StatusOK,
		Data:       books,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleError(w http.ResponseWriter, err error, statusCode int, message string) {
	response := api.Response{
		Type:       "error",
		StatusCode: statusCode,
		Message:    message + "\n" + err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
