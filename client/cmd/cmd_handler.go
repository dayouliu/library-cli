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

// sendGetRequestURLParams sends a GET request to the server and returns the response
func sendGetRequestURLParams(endpoint string, params url.Values) (api.Response, error) {
	resp, err := http.Get(ServerUrl + endpoint + "?" + params.Encode())
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

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

// sendPostRequestURLParams sends a POST request to the server and returns the response using URL params
func sendPostRequestURLParams(endpoint string, params url.Values) (api.Response, error) {
	resp, err := http.Post(ServerUrl+endpoint+"?"+params.Encode(), "", nil)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

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

// sendPostRequestJSONBody sends a POST request to the server and returns the response using a JSON body
func sendPostRequestJSONBody(endpoint string, data interface{}) (api.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return api.Response{}, err
	}

	resp, err := http.Post(ServerUrl+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, err
	}

	var response api.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(string(body))
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

	response, err := sendGetRequestURLParams("/book/list", params)
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

	resp, err := sendPostRequestJSONBody("/book/create", book)
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

	resp, err := sendPostRequestJSONBody("/book/set", book)
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

	resp, err := sendPostRequestURLParams("/book/remove", params)
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
		response, err := sendGetRequestURLParams("/collection/list", url.Values{})
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return prettyPrintResponse(response, true, "")
	} else {
		collectionName := args[0]

		params := url.Values{}
		params.Set("collection_name", collectionName)

		resp, err := sendGetRequestURLParams("/collection/list/books", params)

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
	resp, err := sendPostRequestURLParams("/collection/create", params)

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
	resp, err := sendPostRequestURLParams("/collection/remove", params)

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
	resp, err := sendPostRequestURLParams("/collection/add-book", params)

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
	resp, err := sendPostRequestURLParams("/collection/remove-book", params)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return prettyPrintResponse(resp, false, resp.Message)
}
