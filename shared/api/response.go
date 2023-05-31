package api

import "time"

var PublishTimeLayoutDMY = "02/01/2006"

type Book struct {
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	Edition     string    `json:"edition"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
}

type Collection struct {
	Name  string   `json:"name"`
	Books []string `json:"books"`
}

type Response struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
}
