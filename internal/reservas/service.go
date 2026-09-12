package reservas

type Service struct {
	store ReservaStore
}

func NewService(store ReservaStore) *Service {
	return &Service{store: store}
}

func (s *Service) Reservar(r Reserva) error {
	if r.FilmeID == "" {
		return ErrFilmeIDVazio
	}
	if r.AssentoID == "" {
		return ErrAssentoIDVazio
	}
	if r.UsuarioID == "" {
		return ErrUsuarioIDVazio
	}

	return s.store.Reservar(r)
}

func (s *Service) ListarReservas(filmeID string) []Reserva {
	return s.store.ListarReservas(filmeID)
}
