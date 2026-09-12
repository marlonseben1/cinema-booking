package mensageria

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marlonseben/cinema-booking/internal/reservas"
)

const FilaComandos = "reservas.comandos"

const replyTo = "amq.rabbitmq.reply-to"

type RPCClient struct {
	ch      *amqp.Channel
	timeout time.Duration

	mu      sync.Mutex
	pending map[string]chan reservas.RespostaComando
}

func NewRPCClient(conn *amqp.Connection, timeout time.Duration) (*RPCClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("abrir canal: %w", err)
	}

	if _, err := ch.QueueDeclare(FilaComandos, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declarar fila de comandos: %w", err)
	}

	msgs, err := ch.Consume(replyTo, "", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consumir respostas via reply-to: %w", err)
	}

	c := &RPCClient{
		ch:      ch,
		timeout: timeout,
		pending: make(map[string]chan reservas.RespostaComando),
	}

	go c.consumirRespostas(msgs)

	return c, nil
}

func (c *RPCClient) consumirRespostas(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		c.mu.Lock()
		respCh, ok := c.pending[d.CorrelationId]
		if ok {
			delete(c.pending, d.CorrelationId)
		}
		c.mu.Unlock()

		if !ok {
			continue
		}

		var resp reservas.RespostaComando
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			resp = reservas.RespostaComando{Erro: "resposta inválida do worker"}
		}
		respCh <- resp
	}
}

func (c *RPCClient) Publicar(ctx context.Context, cmd reservas.ComandoReservarAssento) (reservas.RespostaComando, error) {
	body, err := json.Marshal(cmd)
	if err != nil {
		return reservas.RespostaComando{}, fmt.Errorf("serializar comando: %w", err)
	}

	corrID := uuid.NewString()
	respCh := make(chan reservas.RespostaComando, 1)

	c.mu.Lock()
	c.pending[corrID] = respCh
	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err = c.ch.PublishWithContext(ctx, "", FilaComandos, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		ReplyTo:       replyTo,
		Body:          body,
	})
	if err != nil {
		c.mu.Lock()
		delete(c.pending, corrID)
		c.mu.Unlock()
		return reservas.RespostaComando{}, fmt.Errorf("publicar comando: %w", err)
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, corrID)
		c.mu.Unlock()
		return reservas.RespostaComando{}, fmt.Errorf("timeout aguardando resposta do worker: %w", ctx.Err())
	}
}
