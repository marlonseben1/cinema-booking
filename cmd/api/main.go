package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/mensageria"
	"github.com/marlonseben/cinema-booking/internal/reservas"
	httptransport "github.com/marlonseben/cinema-booking/internal/transport/http"
)

const timeoutRPC = 5 * time.Second

// @title			Cinema Booking API
// @version		1.0
// @description	API para reserva de assentos de cinema, com locking pessimista distribuído via RabbitMQ.
// @BasePath		/
func main() {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgres://cinema:cinema@localhost:5432/cinema_booking?sslmode=disable"
	}

	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		log.Fatalf("conectar ao postgres: %v", err)
	}
	defer db.Close()

	store := reservas.NewPostgresStore(db)
	if err := store.Migrate(); err != nil {
		log.Fatalf("aplicar migrations: %v", err)
	}

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatalf("conectar ao rabbitmq: %v", err)
	}
	defer conn.Close()

	publisher, err := mensageria.NewRPCClient(conn, timeoutRPC)
	if err != nil {
		log.Fatalf("criar rpc client: %v", err)
	}

	service := reservas.NewService(store, publisher)
	router := httptransport.NewRouter(service)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
