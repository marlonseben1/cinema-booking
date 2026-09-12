package reservas

import (
	"context"
	"errors"
)

var (
	ErrAssentoOcupado = errors.New("assento já reservado")
	ErrFilmeIDVazio   = errors.New("filme id não pode ser vazio")
	ErrAssentoIDVazio = errors.New("assento id não pode ser vazio")
	ErrUsuarioIDVazio = errors.New("usuario id não pode ser vazio")
)

type Reserva struct {
	ID        string
	FilmeID   string
	AssentoID string
	UsuarioID string
	Status    string
}

type ReservaStore interface {
	Reservar(r Reserva) error
	ListarReservas(filmeID string) []Reserva
}

type ComandoReservarAssento struct {
	FilmeID   string
	AssentoID string
	UsuarioID string
}

type RespostaComando struct {
	Sucesso bool
	Erro    string
}

type ComandoPublisher interface {
	Publicar(ctx context.Context, cmd ComandoReservarAssento) (RespostaComando, error)
}
