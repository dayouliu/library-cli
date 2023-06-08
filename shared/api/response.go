package api

import "time"

var PublishTimeLayoutDMY = "2006-01-02"

type Book struct {
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	PublishDate time.Time `json:"publish_date"`
	Edition     string    `json:"edition"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
}

type Response struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
}
