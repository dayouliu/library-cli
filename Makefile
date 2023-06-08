# Build command
cli:
	go build -o bms client/main.go

db:
	docker-compose down && docker-compose up -d

api:
	go run server/main.go