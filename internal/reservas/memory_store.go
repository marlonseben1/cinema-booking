package reservas

import (
	"sync"
)

type MemoryStore struct {
	mu       sync.Mutex
	reservas map[string]Reserva
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		reservas: map[string]Reserva{},
	}
}

func (s *MemoryStore) Reservar(r Reserva) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chave := r.FilmeID + "|" + r.AssentoID
	if _, ocupado := s.reservas[chave]; ocupado {
		return ErrAssentoOcupado
	}

	s.reservas[chave] = r
	return nil
}

func (s *MemoryStore) ListarReservas(filmeID string) []Reserva {
	s.mu.Lock()
	defer s.mu.Unlock()

	resultado := []Reserva{}
	for _, r := range s.reservas {
		if r.FilmeID == filmeID {
			resultado = append(resultado, r)
		}
	}

	return resultado
}
