package reservas

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
