# Getting started

The book management system (bms) is broken into 3 parts:

- the CLI client located in the client directory
- the server serving REST APIs located in the server directory running on port `8080`
- Postgres database storing the book and collection data.

### Running the app

The Postgres database can be started by running the Docker Compose file

```
docker-compose up -d
```

Alternatively, you can create the Postgres database manually with the following config:

```
port: 5432
POSTGRES_USER: postgres
POSTGRES_PASSWORD: password
POSTGRES_DB: bms_db
```

To run the server on port 8080 (within the project root directory):

```
go run server/main.go
```

To build the CLI client and generate the `bms` binary

```
go build -o bms client/main.go
```

Run tests
```
go test ./tests -v
```

The code is tested with the following go version:

```
❯ go version
go version go1.20.4 darwin/amd64
```

# Step 1: User Experience CLI Client

- commands follow POSIX conventions
- some conventions are inspired by LXD

### Running the client

To run the client and list commands:

```
./bms --help
```

Alternatively, you can use

```
go run client/main.go --help
```

### Create a book

```
./bms book create "book title 1" --author="author 1" --description="description 1" --genre="mystery" --publish_date="2000-01-01" --edition="1"
```

- Only the title is required for creating a book (passed in as a command argument). All flag arguments are optional and will have a default value if not initialized
- Date time format for `publish_date` should be in the form `YYYY-MM-DD`

### Set book attributes

```
./bms book set "book title 1" --author="author 1 update" --description="description 1 update" --genre="adventure" --publish_date="2000-01-02" --edition="2"
```

- Only the title is required for creating a book (passed in as a command argument). All flag arguments are optional.
- Date time format for `publish_date` should be in the form `YYYY-MM-DD`

### List books

List all books (with optional filters)

```bash
./bms book list # list all books
./bms book list --title="book 1" # get info of "book 1"
./bms book list --genre="adventure" --publish_start="2000-01-01" --publish_end="2000-02-01" # filter books by genre and publish date range
```

- All filter flags are optional and order does not matter
- Date time format for `publish_start` and `publish_end` should be in the form `YYYY-MM-DD`

Sample command output:
```
[
 {
  "author": "author1",
  "description": "description1",
  "edition": "1",
  "genre": "genre1",
  "publish_date": "2000-01-01T00:00:00Z",
  "title": "book1"
 },
 {
  "author": "author2",
  "description": "description2",
  "edition": "2",
  "genre": "genre2",
  "publish_date": "2000-01-02T00:00:00Z",
  "title": "book2"
 },
 {
  "author": "author3",
  "description": "description3",
  "edition": "3",
  "genre": "genre3",
  "publish_date": "2000-01-03T00:00:00Z",
  "title": "book3"
 }
]
```

### Remove book

```bash
./bms book remove "book title"
```

### Create collection

```bash
./bms collection create "collection 1"
```

### Add book to collection

```bash
./bms collection add-book "collection 1" "book 1"
```

### Remove book from collection

```bash
./bms collection remove-book "collection 1" "book 1"
```

### List books in a collection

```bash
./bms collection list "collection 1"
```

### List all collections

```bash
./bms collection list
```

### Remove collection

```bash
./bms collection remove "collection 1"
```

# Step 2: REST API Server

### Structure

Sample success response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book updated successfully",
    "data": null
}
```

Sample error response:

```bash
{
    "type": "error",
    "status_code": 500,
    "message": "Error creating collection\npq: duplicate key value violates unique constraint \"collections_pkey\"",
    "data": null
}
```

### Create book endpoint

`book/create`

- POST request with JSON request body

Example JSON request body:

```bash
{
	"title": "The Lord of the Rings",
	"author": "J.R.R. Tolkien",
	"publish_date": "1954-07-29",
	"edition": "1st",
	"description": "The Lord of the Rings is an epic high-fantasy novel written by English author and scholar J. R. R. Tolkien.",
	"genre": "Fantasy"
}
```

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book created successfully",
    "data": null
}
```

### Set book endpoint

`book/set`

- POST request with JSON request body

Example JSON request body:

```bash
{
	"title": "The Lord of the Rings",
	"author": "J.R.R. Tolkien",
	"publish_date": "1954-07-29",
	"edition": "1st",
	"description": "The Lord of the Rings is an epic high-fantasy novel written by English author and scholar J. R. R. Tolkien.",
	"genre": "Fantasy"
}
```

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book updated successfully",
    "data": null
}
```

### Remove book endpoint

`book/remove`

- POST request with title URL parameter

Example request:

- `localhost:8080/book/remove?title="book 1"`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book removed successfully",
    "data": null
}
```

### List book endpoint

`book/list`

- GET request with URL filter parameters (`author`, `genre`, `publish_start`, `publish_end`)
- `publish_start`, `publish_end` must be in `YYYY-MM-DD` format and filters books in the range `[publish_start, publish_end]` inclusive where `publish_start < publish_end`
- All filter parameters are optional, all books are returned if no filters are provided

Example request:

- `localhost:8080/book/list?title="book 1"`
- `localhost:8080/book/list?author=author1&genre=mystery&publish_start=2000-01-31&publish_end=2000-03-31`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Books retrieved successfully",
    "data": [
        {
            "title": "book1",
            "author": "author1",
            "publish_date": "2000-01-01T00:00:00Z",
            "edition": "1",
            "description": "description1",
            "genre": "genre1"
        },
        {
            "title": "book2",
            "author": "author2",
            "publish_date": "2000-01-02T00:00:00Z",
            "edition": "1",
            "description": "description2",
            "genre": "genre2"
        },
        {
            "title": "book3",
            "author": "author3",
            "publish_date": "2000-01-03T00:00:00Z",
            "edition": "1",
            "description": "description3",
            "genre": "genre3"
        }
    ]
}
```

### Create collection endpoint

`collection/create`

- POST request with required `collection_name` URL parameter

Example request:

- `collection/create?collection_name="collection 1"`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Collection created successfully",
    "data": null
}
```

### List collection endpoint

`collection/list`

- GET request

Example request:

- `collection/list`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Collections retrieved successfully",
    "data": [
        "collection1",
        "collection2",
        "collection3"
    ]
}
```

### List books in collection endpoint

`collection/list/books`

- GET request with required `collection_name` URL parameter

Example request:

- `localhost:8080/collection/list/books?collection_name=collection1`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Books in collection retrieved successfully",
    "data": [
        "book1"
    ]
}
```

### Remove collection endpoint

`collection/remove`

- GET request with required `collection_name` URL parameter

Example request:

- `localhost:8080/collection/remove?collection_name=collection3`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Collection removed successfully",
    "data": null
}
```

### Add book to collection

`collection/add-book`

- POST request with required `collection_name` and `book_title` URL parameter

Example request:

- `localhost:8080/collection/add-book?collection_name=collection1&book_title=book1`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book added to collection successfully",
    "data": null
}
```

### Remove book from collection

`collection/remove-book`

- POST request with required `collection_name` and `book_title` URL parameter

Example request:

- `localhost:8080/collection/remove-book?collection_name=collection1&book_title=book1`

Example JSON response:

```bash
{
    "type": "success",
    "status_code": 201,
    "message": "Book removed from collection successfully",
    "data": null
}
```

# Step 3: SQL Database

```
CREATE TABLE IF NOT EXISTS books (
    title VARCHAR(255) NOT NULL PRIMARY KEY,
    author VARCHAR(255),
    publish_date DATE,
    edition VARCHAR(10),
    description TEXT,
    genre VARCHAR(255)
);
```

```
CREATE TABLE IF NOT EXISTS collections (
    name VARCHAR(255) NOT NULL PRIMARY KEY,
    description TEXT
);
```

```
CREATE TABLE IF NOT EXISTS collection_subscriptions (
    book_title VARCHAR(255),
    collection_name VARCHAR(255),
    PRIMARY KEY (book_title, collection_name),
    FOREIGN KEY (book_title) REFERENCES books (title), 
    FOREIGN KEY (collection_name) REFERENCES collections (name)
);
```