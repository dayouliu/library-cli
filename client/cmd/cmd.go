package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "bms",
	Short: "Book management CLI",
}

var bookCmd = &cobra.Command{
	Use:   "book",
	Short: "Commands involving books",
}

var listBookCmd = &cobra.Command{
	Use:   "list",
	Short: "List books",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(listBooks(cmd, args))
	},
}

var createBookCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a book",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(createBook(cmd, args))
	},
}

var setBookCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a book",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(setBook(cmd, args))
	},
}

var removeBookCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a book",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(removeBook(cmd, args))
	},
}

var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Commands involving collections",
}

var createCollectionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a collection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(createCollection(cmd, args))
	},
}

var removeCollectionCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a collection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(removeCollection(cmd, args))
	},
}

var addBookToCollectionCmd = &cobra.Command{
	Use:   "add-book",
	Short: "Add a book to a collection",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(addBookToCollection(cmd, args))
	},
}

var removeBookFromCollectionCmd = &cobra.Command{
	Use:   "remove-book",
	Short: "Remove a book from a collection",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(removeBookFromCollection(cmd, args))
	},
}

var listCollectionCmd = &cobra.Command{
	Use:   "list",
	Short: "List books in a collection",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(listCollection(cmd, args))
	},
}

func init() {
	// optional args for createBookCmd
	createBookCmd.Flags().StringP("title", "", "", "Title of the book")
	createBookCmd.Flags().StringP("author", "", "", "Author of the book")
	createBookCmd.Flags().StringP("genre", "", "", "Genre of the book")
	createBookCmd.Flags().StringP("publish_date", "", "", "publish date of the book")
	createBookCmd.Flags().StringP("description", "", "", "Description of the book")
	createBookCmd.Flags().StringP("edition", "", "", "Edition of the book")

	// optional args for listBookCmd

	listBookCmd.Flags().StringP("title", "", "", "Get book with title")
	listBookCmd.Flags().StringP("author", "", "", "Filter books by author")
	listBookCmd.Flags().StringP("genre", "", "", "Filter books by genre")
	listBookCmd.Flags().StringP("publish_start", "", "", "Filter books from publish start date (YYYY-MM-DD)")
	listBookCmd.Flags().StringP("publish_end", "", "", "Filter books to publish end date (YYYY-MM-DD)")

	// optional args for setBookCmd
	setBookCmd.Flags().StringP("author", "", "", "Author of the book")
	setBookCmd.Flags().StringP("genre", "", "", "Genre of the book")
	setBookCmd.Flags().StringP("publish_date", "", "", "publish date of the book (YYYY-MM-DD)")
	setBookCmd.Flags().StringP("description", "", "", "Description of the book")
	setBookCmd.Flags().StringP("edition", "", "", "Edition of the book")

	// book subcommands
	bookCmd.AddCommand(listBookCmd)
	bookCmd.AddCommand(createBookCmd)
	bookCmd.AddCommand(setBookCmd)
	bookCmd.AddCommand(removeBookCmd)

	// collection subcommands
	collectionCmd.AddCommand(createCollectionCmd)
	collectionCmd.AddCommand(addBookToCollectionCmd)
	collectionCmd.AddCommand(removeBookFromCollectionCmd)
	collectionCmd.AddCommand(listCollectionCmd)
	collectionCmd.AddCommand(removeCollectionCmd)

	// root subcommands
	RootCmd.AddCommand(bookCmd)
	RootCmd.AddCommand(collectionCmd)
}
