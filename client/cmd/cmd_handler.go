package cmd

import (
	"bms/shared/api"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	ServerUrl = "http://localhost:8080"
)

func makeRequest(method string, endpoint string, params url.Values, payload interface{}) (api.Response, error) {
	// Create the URL with query parameters
	requestURL, err := url.Parse(ServerUrl + endpoint)
	if err != nil {
		return api.Response{}, err
	}
	requestURL.RawQuery = params.Encode()

	// Convert payload to JSON
	var payloadBytes []byte
	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return api.Response{}, err
		}
	}

	// Create the HTTP request
	request, err := http.NewRequest(method, requestURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return api.Response{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, err
	}

	var response api.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return api.Response{}, err
	}

	return response, nil
}

// prettyPrintResponse formats and prints the response for the cli
func prettyPrintResponse(response api.Response, printJson bool, successMessage string) string {
	if response.Type == "error" {
		return fmt.Sprintf("Error: %s", response.Message)
	}

	jsonData, err := json.MarshalIndent(response.Data, "", " ")
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	if printJson {
		return fmt.Sprintf("%s", string(jsonData))
	}

	if successMessage != "" {
		return successMessage
	}

	return ""
}

// listBooks lists all books in system
func listBooks(cmd *cobra.Command, args []string) string {
	title, _ := cmd.Flags().GetString("title")
	author, _ := cmd.Flags().GetString("author")
	genre, _ := cmd.Flags().GetString("genre")
	publishDateStart, _ := cmd.Flags().GetString("publish_start")
	publishDateEnd, _ := cmd.Flags().GetString("publish_end")

	params := url.Values{}
	if title != "" {
		params.Add("title", title)
	}
	if author != "" {
		params.Add("author", author)
	}
	if genre != "" {
		params.Add("genre", genre)
	}
	if publishDateStart != "" {
		params.Add("publish_start", publishDateStart)
	}
	if publishDateEnd != "" {
		params.Add("publish_end", publishDateEnd)
	}

	response, err := makeRequest(http.MethodGet, "/book/list", params, nil)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(response, true, "")
}

// createBook creates a new book
func createBook(cmd *cobra.Command, args []string) string {
	title := args[0]
	author, _ := cmd.Flags().GetString("author")
	genre, _ := cmd.Flags().GetString("genre")
	publishDateStr, _ := cmd.Flags().GetString("publish_date")
	description, _ := cmd.Flags().GetString("description")
	edition, _ := cmd.Flags().GetString("edition")

	var publishDate time.Time
	var err error
	if publishDateStr != "" {
		publishDate, err = time.Parse(api.PublishTimeLayoutDMY, publishDateStr)
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
	}

	book := api.Book{
		Title:       title,
		Author:      author,
		Genre:       genre,
		PublishDate: publishDate,
		Description: description,
		Edition:     edition,
	}

	resp, err := makeRequest(http.MethodPost, "/book/create", nil, book)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// setBook sets a book's attributes optionally given the book title
func setBook(cmd *cobra.Command, args []string) string {
	title := args[0]
	author, _ := cmd.Flags().GetString("author")
	genre, _ := cmd.Flags().GetString("genre")
	publishDateStr, _ := cmd.Flags().GetString("publish_date")
	description, _ := cmd.Flags().GetString("description")
	edition, _ := cmd.Flags().GetString("edition")

	var publishDate time.Time
	var err error
	if publishDateStr != "" {
		publishDate, err = time.Parse(api.PublishTimeLayoutDMY, publishDateStr)
		if err != nil {
			fmt.Println(publishDateStr)
			return fmt.Sprintf("Error: %s", err)
		}
	}

	book := api.Book{
		Title:       title,
		Author:      author,
		Genre:       genre,
		PublishDate: publishDate,
		Description: description,
		Edition:     edition,
	}

	resp, err := makeRequest(http.MethodPut, "/book/set", nil, book)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// removeBook removes a book from the system
func removeBook(cmd *cobra.Command, args []string) string {
	title := args[0]

	params := url.Values{}
	params.Set("title", title)

	resp, err := makeRequest(http.MethodDelete, "/book/remove", params, nil)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// listCollection either:
// list all collections if collection_name arg is not provided
// list all books in collection_name arg if arg is provided
func listCollection(cmd *cobra.Command, args []string) string {
	if len(args) == 0 {
		response, err := makeRequest(http.MethodGet, "/collection/list", nil, nil)
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return prettyPrintResponse(response, true, "")
	} else {
		collectionName := args[0]

		params := url.Values{}
		params.Set("collection_name", collectionName)

		resp, err := makeRequest(http.MethodGet, "/collection/list/books", params, nil)

		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}

		return prettyPrintResponse(resp, true, "")
	}
}

// createCollection creates a new collection
func createCollection(cmd *cobra.Command, args []string) string {
	collectionName := args[0]

	// post request with url parameters
	params := url.Values{}
	params.Set("collection_name", collectionName)
	resp, err := makeRequest(http.MethodPost, "/collection/create", params, nil)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// removeCollection removes a collection
func removeCollection(cmd *cobra.Command, args []string) string {
	collectionName := args[0]

	// post request with url parameters
	params := url.Values{}
	params.Set("collection_name", collectionName)
	resp, err := makeRequest(http.MethodDelete, "/collection/remove", params, nil)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// addBookToCollection adds a book to a collection
func addBookToCollection(cmd *cobra.Command, args []string) string {
	collectionName := args[0]
	bookTitle := args[1]

	// post request with url parameters
	params := url.Values{}
	params.Set("collection_name", collectionName)
	params.Set("book_title", bookTitle)
	resp, err := makeRequest(http.MethodPost, "/collection/add-book", params, nil)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}

// removeBookFromCollection removes a book from a collection
func removeBookFromCollection(cmd *cobra.Command, args []string) string {
	collectionName := args[0]
	bookTitle := args[1]

	// post request with url params
	params := url.Values{}
	params.Set("collection_name", collectionName)
	params.Set("book_title", bookTitle)
	resp, err := makeRequest(http.MethodDelete, "/collection/remove-book", params, nil)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}
