package tests

import (
	"bms/shared/api"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// readJsonFile reads a JSON file and returns the byte contents
func readJsonFile(filename string) ([]byte, error) {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// mockRespondError mocks a server error response
func mockRespondError(w http.ResponseWriter, err error, statusCode int, message string) {
	errorMessage := message
	if err != nil {
		errorMessage = fmt.Sprintf("%s\n%s", message, err)
	}

	response := api.Response{
		Type:       "error",
		StatusCode: statusCode,
		Message:    errorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// mockRespondJSON mocks a server successful JSON response
func mockRespondJSON(w http.ResponseWriter, data interface{}) {
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

// mockRouter mocks a router for testing
func mockRouter(w http.ResponseWriter, r *http.Request) {
	// handle book/create route
	if r.Method == "POST" && r.URL.Path == "/book/create" {
		mockCreateBook(w, r)
	} else if r.Method == "GET" && r.URL.Path == "/book/list" {
		mockListBooks(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// mockCreateBook mocks the book/create route
func mockCreateBook(w http.ResponseWriter, r *http.Request) {
	// add book
	var book api.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		mockRespondError(w, err, http.StatusBadRequest, "Invalid request body")
		return
	}

	if book.Title == "" {
		mockRespondError(w, err, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	if err != nil {
		mockRespondError(w, err, http.StatusInternalServerError, "Error creating book")
		return
	}

	mockRespondJSON(w, nil)
}

// mockListBooks mocks the book/list route
func mockListBooks(w http.ResponseWriter, r *http.Request) {
	// load mock_book_list.json file in current directory
	var books []api.Book

	// Unmarshal the JSON data into the books variable
	data, err := readJsonFile("resources/mock_books.json")
	err = json.Unmarshal(data, &books)
	if err != nil {
		mockRespondError(w, err, http.StatusInternalServerError, "Error listing books")
		return
	}

	mockRespondJSON(w, books)
}
