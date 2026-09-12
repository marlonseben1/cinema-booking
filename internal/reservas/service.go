package reservas

import (
	"context"
	"errors"
)

type Service struct {
	store     ReservaStore
	publisher ComandoPublisher
}

func NewService(store ReservaStore, publisher ComandoPublisher) *Service {
	return &Service{store: store, publisher: publisher}
}

func (s *Service) Reservar(ctx context.Context, r Reserva) error {
	if err := ValidarReserva(r); err != nil {
		return err
	}

	resp, err := s.publisher.Publicar(ctx, ComandoReservarAssento{
		FilmeID:   r.FilmeID,
		AssentoID: r.AssentoID,
		UsuarioID: r.UsuarioID,
	})
	if err != nil {
		return err
	}

	if !resp.Sucesso {
		if resp.Erro == ErrAssentoOcupado.Error() {
			return ErrAssentoOcupado
		}
		return errors.New(resp.Erro)
	}

	return nil
}

func (s *Service) ListarReservas(filmeID string) []Reserva {
	return s.store.ListarReservas(filmeID)
}
