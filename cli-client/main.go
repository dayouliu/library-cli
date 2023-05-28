package main

import (
	"bms/shared/api"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

var (
	authorFilter string
	genreFilter  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bms",
		Short: "Book management CLI",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List books",
		Run:   listBooks,
	}

	listCmd.Flags().StringVar(&authorFilter, "author", "", "Filter books by author")
	listCmd.Flags().StringVar(&genreFilter, "genre", "", "Filter books by genre")

	rootCmd.AddCommand(listCmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func listBooks(cmd *cobra.Command, args []string) {
	fmt.Println("List of books:")
	if authorFilter != "" {
		fmt.Printf("Author filter: %s\n", authorFilter)
	}
	if genreFilter != "" {
		fmt.Printf("Genre filter: %s\n", genreFilter)
	}

	url := "http://localhost:8080/book/list"

	// Send GET request to list books
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Errorf("received non-OK status code: %d", resp.StatusCode))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal JSON response into slice of Book structs
	var response api.Response
	var books []api.Book

	print(string(body))

	err = json.Unmarshal(body, &response)
	books = response.Data.([]api.Book)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Print each book
	for _, book := range books {
		fmt.Println(book)
	}
}
