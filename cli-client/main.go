package main

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
	SERVER_URL         = "http://localhost:8080"
	BOOK_LIST_ENDPOINT = "/book/list"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bms",
		Short: "Book management CLI",
	}

	bookCmd := &cobra.Command{
		Use:   "book",
		Short: "commands involving books",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list books",
		Run:   listBooks,
	}
	listCmd.Flags().StringP("author", "", "", "Filter books by author")
	listCmd.Flags().StringP("genre", "", "", "Filter books by genre")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a book",
		Args:  cobra.ExactArgs(1),
		Run:   createBook,
	}
	createCmd.Flags().StringP("title", "", "", "Title of the book")
	createCmd.Flags().StringP("author", "", "", "Author of the book")
	createCmd.Flags().StringP("genre", "", "", "Genre of the book")
	createCmd.Flags().StringP("published_at", "", "", "Published date of the book")
	createCmd.Flags().StringP("description", "", "", "Description of the book")
	createCmd.Flags().StringP("edition", "", "", "Edition of the book")

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a book",
		Args:  cobra.ExactArgs(1),
		Run:   setBook,
	}
	setCmd.Flags().StringP("author", "a", "", "Author of the book")
	setCmd.Flags().StringP("genre", "g", "", "Genre of the book")
	setCmd.Flags().StringP("published_at", "p", "", "Published date of the book")
	setCmd.Flags().StringP("description", "d", "", "Description of the book")
	setCmd.Flags().StringP("edition", "e", "", "Edition of the book")

	collectionCmd := &cobra.Command{
		Use:   "collection",
		Short: "commands involving collections",
	}

	createCollectionCmd := &cobra.Command{
		Use:   "create",
		Short: "create a collection",
		Args:  cobra.ExactArgs(1),
		Run:   createCollection,
	}

	addCollectionCmd := &cobra.Command{
		Use:   "add",
		Short: "add a book to a collection",
		Args:  cobra.ExactArgs(2),
		Run:   addToCollection,
	}

	removeCollectionCmd := &cobra.Command{
		Use:   "remove",
		Short: "remove a book from a collection",
		Args:  cobra.ExactArgs(2),
		Run:   removeFromCollection,
	}

	listCollectionBooksCmd := &cobra.Command{
		Use:   "list",
		Short: "list books in a collection",
		Run:   listCollections,
	}

	bookCmd.AddCommand(listCmd)
	bookCmd.AddCommand(createCmd)
	bookCmd.AddCommand(setCmd)

	collectionCmd.AddCommand(createCollectionCmd)
	collectionCmd.AddCommand(addCollectionCmd)
	collectionCmd.AddCommand(removeCollectionCmd)
	collectionCmd.AddCommand(listCollectionBooksCmd)

	rootCmd.AddCommand(bookCmd)
	rootCmd.AddCommand(collectionCmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func sendGetRequest(endpoint string) (api.Response, error) {
	// Send GET request to list books
	fmt.Println("Sending GET request to " + SERVER_URL + endpoint)

	resp, err := http.Get(SERVER_URL + endpoint)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	fmt.Println(resp)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, err
	}

	// Unmarshal JSON response into slice of Book structs
	var response api.Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		return api.Response{}, err
	}

	return response, nil
}

func sendPostRequest(endpoint string, params url.Values) (api.Response, error) {
	fmt.Println("Sending POST request to " + SERVER_URL + endpoint + "?" + params.Encode())

	resp, err := http.Post(SERVER_URL+endpoint+"?"+params.Encode(), "", nil)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, err
	}

	// Unmarshal JSON response into slice of Book structs
	var response api.Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		return api.Response{}, err
	}

	return response, nil
}

func sendPostRequestJSONBody(endpoint string, data interface{}) (api.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return api.Response{}, err
	}

	resp, err := http.Post(SERVER_URL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, err
	}

	// Unmarshal JSON response into slice of Book structs
	var response api.Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		return api.Response{}, err
	}

	return response, nil
}

func prettyPrintResponse(response api.Response, printJson bool, successMessage string) {
	if response.Type == "error" {
		fmt.Println(response.Message)
		return
	}

	jsonData, err := json.MarshalIndent(response.Data, "", " ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print each book
	if printJson {
		fmt.Println(string(jsonData))
	}

	if successMessage != "" {
		fmt.Println(successMessage)
	}
}

func listBooks(cmd *cobra.Command, args []string) {
	response, err := sendGetRequest(BOOK_LIST_ENDPOINT)
	if err != nil {
		fmt.Println(err)
		return
	}

	prettyPrintResponse(response, true, "")
}

func createBook(cmd *cobra.Command, args []string) {
	title := args[0]
	author, _ := cmd.Flags().GetString("author")
	genre, _ := cmd.Flags().GetString("genre")
	publishedAtStr, _ := cmd.Flags().GetString("published_at")
	description, _ := cmd.Flags().GetString("description")
	edition, _ := cmd.Flags().GetString("edition")

	// convert UTC string to time.Time
	var publishedAt time.Time
	var err error
	if publishedAtStr != "" {
		publishedAt, err = time.Parse(api.PublishTimeLayoutDMY, publishedAtStr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	book := api.Book{
		Title:       title,
		Author:      author,
		Genre:       genre,
		PublishedAt: publishedAt,
		Description: description,
		Edition:     edition,
	}

	resp, err := sendPostRequestJSONBody("/book/create", book)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Book created successfully")
}

func setBook(cmd *cobra.Command, args []string) {
	// convert UTC string to time.Time

	title := args[0]
	author, _ := cmd.Flags().GetString("author")
	genre, _ := cmd.Flags().GetString("genre")
	publishedAtStr, _ := cmd.Flags().GetString("published_at")
	description, _ := cmd.Flags().GetString("description")
	edition, _ := cmd.Flags().GetString("edition")

	var publishedAt time.Time
	var err error
	if publishedAtStr != "" {
		publishedAt, err = time.Parse(api.PublishTimeLayoutDMY, publishedAtStr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	book := api.Book{
		Title:       title,
		Author:      author,
		Genre:       genre,
		PublishedAt: publishedAt,
		Description: description,
		Edition:     edition,
	}

	resp, err := sendPostRequestJSONBody("/book/set", book)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Book created successfully")
}

func listCollections(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("!!!")
		response, err := sendGetRequest("/collection/list")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("???")
		prettyPrintResponse(response, true, "")
	} else {
		collectionName := args[0]

		params := url.Values{}
		params.Set("collection_name", collectionName)

		// post request with url parameters

		resp, err := sendPostRequest("/collection/list/books", params)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		prettyPrintResponse(resp, true, "")
	}
}

func createCollection(cmd *cobra.Command, args []string) {
	collectionName := args[0]

	params := url.Values{}
	fmt.Println("collectionName", collectionName)
	params.Set("collection_name", collectionName)

	// post request with url parameters

	resp, err := sendPostRequest("/collection/create", params)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Collection created successfully")
}

func addToCollection(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	bookTitle := args[1]

	params := url.Values{}
	params.Set("collection_name", collectionName)
	params.Set("book_title", bookTitle)

	// post request with url parameters

	resp, err := sendPostRequest("/collection/add", params)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Book added to collection successfully")
}

func removeFromCollection(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	bookTitle := args[1]

	params := url.Values{}
	params.Set("collection_name", collectionName)
	params.Set("book_title", bookTitle)

	// post request with url parameters

	resp, err := sendPostRequest("/collection/remove", params)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Book added to collection successfully")
}
