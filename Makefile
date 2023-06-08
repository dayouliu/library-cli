# Build command
cli:
	go build -o bms client/main.go

database:
	docker-compose down && docker-compose up -d
