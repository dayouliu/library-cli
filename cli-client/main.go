package main

import (
	"bms/shared/api"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"time"
)

var (
	SERVER_URL         = "http://localhost:8080"
	BOOK_LIST_ENDPOINT = "/book/list"
	authorFilter       string
	genreFilter        string
)

var (
	titleFlag       string
	authorFlag      string
	genreFlag       string
	publishedAtFlag string
	descriptionFlag string
	editionFlag     string
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
	listCmd.Flags().StringVar(&authorFilter, "author", "", "Filter books by author")
	listCmd.Flags().StringVar(&genreFilter, "genre", "", "Filter books by genre")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a book",
		Run:   createBook,
	}
	createCmd.Flags().StringVar(&titleFlag, "title", "", "Title of the book")
	createCmd.Flags().StringVar(&authorFlag, "author", "", "Author of the book")
	createCmd.Flags().StringVar(&genreFlag, "genre", "", "Genre of the book")
	createCmd.Flags().StringVar(&publishedAtFlag, "published_at", "", "Published date of the book")
	createCmd.Flags().StringVar(&descriptionFlag, "description", "", "Description of the book")
	createCmd.Flags().StringVar(&editionFlag, "edition", "", "Edition of the book")

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

	bookCmd.AddCommand(listCmd)
	bookCmd.AddCommand(createCmd)
	bookCmd.AddCommand(setCmd)

	rootCmd.AddCommand(bookCmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func sendGetRequest(endpoint string) (api.Response, error) {
	// Send GET request to list books
	resp, err := http.Get(SERVER_URL + endpoint)
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

func sendPostRequest(endpoint string, data interface{}) (api.Response, error) {
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
	if authorFilter != "" {
		fmt.Printf("Author filter: %s\n", authorFilter)
	}
	if genreFilter != "" {
		fmt.Printf("Genre filter: %s\n", genreFilter)
	}

	response, err := sendGetRequest(BOOK_LIST_ENDPOINT)
	if err != nil {
		fmt.Println(err)
		return
	}

	prettyPrintResponse(response, true, "")
}

func createBook(cmd *cobra.Command, args []string) {
	// convert UTC string to time.Time
	var publishedAt time.Time
	var err error
	if publishedAtFlag != "" {
		publishedAt, err = time.Parse(api.PublishTimeLayoutDMY, publishedAtFlag)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	book := api.Book{
		Title:       titleFlag,
		Author:      authorFlag,
		Genre:       genreFlag,
		PublishedAt: publishedAt,
		Description: descriptionFlag,
		Edition:     editionFlag,
	}

	resp, err := sendPostRequest("/book/create", book)
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

	resp, err := sendPostRequest("/book/set", book)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyPrintResponse(resp, false, "Book created successfully")
}
