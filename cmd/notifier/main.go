package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/mensageria"
)

func main() {
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatalf("conectar ao rabbitmq: %v", err)
	}
	defer conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("notifier pronto, consumindo reservas.notificacoes")

	err = mensageria.ConsumirNotificacoes(ctx, conn, func(n mensageria.NotificacaoReserva) {
		log.Printf("notificação enviada para usuário %s: assento %s confirmado no filme %s", n.UsuarioID, n.AssentoID, n.FilmeID)
	})
	if err != nil && err != context.Canceled {
		log.Fatalf("consumir notificações: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
