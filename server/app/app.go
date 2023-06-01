package app

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

type Config struct {
	Host       string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
	ServerPort string
}

// App server struct
type App struct {
	Server *http.Server
	// router
	Router   *chi.Mux
	Database *sql.DB
}

func NewApp(config Config) *App {
	// Open a database connection
	var err error
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.DbPort, config.DbUser, config.DbPassword, config.DbName)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	// create the books and collections tables if they don't exist
	createTables(db)

	// initialize the router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	handler := &Handler{db: db}

	// book endpoints
	router.Post("/book/create", handler.createBook)
	router.Get("/book/list", handler.getBooks)
	router.Post("/book/set", handler.setBook)
	router.Post("/book/remove", handler.removeBook)

	// collection endpoints
	router.Post("/collection/create", handler.createCollection)
	router.Post("/collection/add", handler.addToCollection)
	router.Post("/collection/remove", handler.removeCollection)
	router.Post("/collection/remove-book", handler.removeFromCollection)
	router.Get("/collection/list", handler.getCollections)
	router.Post("/collection/list/books", handler.getBooksInCollection)

	// Start the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServerPort), // Specify the address where your server is running
		Handler: router,
	}

	return &App{
		Server:   server,
		Router:   router,
		Database: db,
	}
}

func (a *App) Run() {
	defer a.Database.Close()
	err := a.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

// createTables creates the books and collections tables if they don't exist
func createTables(db *sql.DB) {
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
