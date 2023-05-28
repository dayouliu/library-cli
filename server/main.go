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
	createBooksTableQuery := `CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL,
		published_at DATE NOT NULL,
		edition VARCHAR(10) NOT NULL,
		description TEXT,
		genre VARCHAR(255) NOT NULL
	);`

	createCollectionsTableQuery := `CREATE TABLE IF NOT EXISTS collections (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
    	description TEXT
	);`

	createCollectionSubscriptions := `CREATE TABLE IF NOT EXISTS collection_subscriptions (
		id SERIAL PRIMARY KEY,
		book_id INTEGER REFERENCES books (id),
		collection_id INTEGER REFERENCES collections (id),
		CONSTRAINT unique_book_collection UNIQUE (book_id, collection_id)
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
		handleError(w, http.StatusBadRequest, "Invalid request body")
		log.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO books (title, author, published_at, edition, description, genre) VALUES ($1, $2, $3, $4, $5, $6)`,
		book.Title, book.Author, book.PublishedAt, book.Edition, book.Description, book.Genre)
	if err != nil {
		log.Fatal(err)
		handleError(w, http.StatusInternalServerError, "Error creating book")
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
	rows, err := db.Query("SELECT id, title, author, published_at, edition, description, genre FROM books")
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Error getting books")
		log.Fatal(err)
	}
	defer rows.Close()

	books := make([]api.Book, 0)

	for rows.Next() {
		var book api.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.PublishedAt, &book.Edition, &book.Description, &book.Genre)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error getting books")
			log.Fatal(err)
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

func handleError(w http.ResponseWriter, statusCode int, message string) {
	response := api.Response{
		Type:       "error",
		StatusCode: statusCode,
		Data:       message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
