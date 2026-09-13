package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/mensageria"
	"github.com/marlonseben/cinema-booking/internal/reservas"
)

func main() {
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	postgresDSN := getEnv("POSTGRES_DSN", "postgres://cinema:cinema@localhost:5432/cinema_booking?sslmode=disable")

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

	notificador, err := mensageria.NewNotificacaoPublisher(conn)
	if err != nil {
		log.Fatalf("criar publisher de notificações: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("worker pronto, consumindo reservas.comandos (prefetch=1)")

	err = mensageria.ConsumirComandos(ctx, conn, func(cmd reservas.ComandoReservarAssento) reservas.RespostaComando {
		r := reservas.Reserva{
			FilmeID:   cmd.FilmeID,
			AssentoID: cmd.AssentoID,
			UsuarioID: cmd.UsuarioID,
			Status:    "confirmada",
		}

		if err := reservas.ValidarReserva(r); err != nil {
			return reservas.RespostaComando{Erro: err.Error()}
		}

		if err := store.Reservar(r); err != nil {
			return reservas.RespostaComando{Erro: err.Error()}
		}

		log.Printf("reserva confirmada: filme=%s assento=%s usuario=%s", r.FilmeID, r.AssentoID, r.UsuarioID)

		notif := mensageria.NotificacaoReserva{
			FilmeID:   r.FilmeID,
			AssentoID: r.AssentoID,
			UsuarioID: r.UsuarioID,
		}
		if err := notificador.Publicar(context.Background(), notif); err != nil {
			log.Printf("publicar notificação: %v", err)
		}

		return reservas.RespostaComando{Sucesso: true}
	})
	if err != nil && err != context.Canceled {
		log.Fatalf("consumir comandos: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
