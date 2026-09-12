package mensageria

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/reservas"
)

type ManipuladorComando func(cmd reservas.ComandoReservarAssento) reservas.RespostaComando

func ConsumirComandos(ctx context.Context, conn *amqp.Connection, handler ManipuladorComando) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("abrir canal: %w", err)
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(FilaComandos, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declarar fila de comandos: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("configurar qos: %w", err)
	}

	msgs, err := ch.Consume(FilaComandos, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consumir fila de comandos: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de mensagens fechado")
			}

			processarComando(ctx, ch, d, handler)
		}
	}
}

func processarComando(ctx context.Context, ch *amqp.Channel, d amqp.Delivery, handler ManipuladorComando) {
	var cmd reservas.ComandoReservarAssento
	if err := json.Unmarshal(d.Body, &cmd); err != nil {
		log.Printf("comando inválido: %v", err)
		_ = d.Nack(false, false)
		return
	}

	resp := handler(cmd)

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("serializar resposta: %v", err)
		_ = d.Nack(false, false)
		return
	}

	if d.ReplyTo != "" {
		err = ch.PublishWithContext(ctx, "", d.ReplyTo, false, false, amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		})
		if err != nil {
			log.Printf("publicar resposta: %v", err)
		}
	}

	_ = d.Ack(false)
}
