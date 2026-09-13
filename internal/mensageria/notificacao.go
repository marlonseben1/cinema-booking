package mensageria

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const FilaNotificacoes = "reservas.notificacoes"

type NotificacaoReserva struct {
	FilmeID   string
	AssentoID string
	UsuarioID string
}

type NotificacaoPublisher struct {
	ch *amqp.Channel
}

func NewNotificacaoPublisher(conn *amqp.Connection) (*NotificacaoPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("abrir canal: %w", err)
	}

	if _, err := ch.QueueDeclare(FilaNotificacoes, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declarar fila de notificações: %w", err)
	}

	return &NotificacaoPublisher{ch: ch}, nil
}

func (p *NotificacaoPublisher) Publicar(ctx context.Context, n NotificacaoReserva) error {
	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("serializar notificação: %w", err)
	}

	return p.ch.PublishWithContext(ctx, "", FilaNotificacoes, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

type ManipuladorNotificacao func(n NotificacaoReserva)

func ConsumirNotificacoes(ctx context.Context, conn *amqp.Connection, handler ManipuladorNotificacao) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("abrir canal: %w", err)
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(FilaNotificacoes, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declarar fila de notificações: %w", err)
	}

	msgs, err := ch.Consume(FilaNotificacoes, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consumir fila de notificações: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de mensagens fechado")
			}

			var n NotificacaoReserva
			if err := json.Unmarshal(d.Body, &n); err != nil {
				log.Printf("notificação inválida: %v", err)
				_ = d.Nack(false, false)
				continue
			}

			handler(n)
			_ = d.Ack(false)
		}
	}
}
