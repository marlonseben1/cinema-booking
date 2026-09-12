package main

import (
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/mensageria"
	"github.com/marlonseben/cinema-booking/internal/reservas"
	httptransport "github.com/marlonseben/cinema-booking/internal/transport/http"
)

const timeoutRPC = 5 * time.Second

func main() {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
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

	store := reservas.NewMemoryStore()
	service := reservas.NewService(store, publisher)
	router := httptransport.NewRouter(service)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
