package reservas

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const codigoViolacaoUnique = "23505"

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Migrate() error {
	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("criar driver de migration: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("carregar migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("inicializar migration: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("aplicar migrations: %w", err)
	}

	return nil
}

func (s *PostgresStore) Reservar(r Reserva) error {
	_, err := s.db.Exec(
		`INSERT INTO reservas (filme_id, assento_id, usuario_id, status) VALUES ($1, $2, $3, $4)`,
		r.FilmeID, r.AssentoID, r.UsuarioID, r.Status,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoViolacaoUnique {
		return ErrAssentoOcupado
	}

	return err
}

func (s *PostgresStore) ListarReservas(filmeID string) []Reserva {
	linhas, err := s.db.Query(
		`SELECT id, filme_id, assento_id, usuario_id, status FROM reservas WHERE filme_id = $1`,
		filmeID,
	)
	if err != nil {
		return []Reserva{}
	}
	defer linhas.Close()

	resultado := []Reserva{}
	for linhas.Next() {
		var r Reserva
		var id int
		if err := linhas.Scan(&id, &r.FilmeID, &r.AssentoID, &r.UsuarioID, &r.Status); err != nil {
			return []Reserva{}
		}
		r.ID = strconv.Itoa(id)
		resultado = append(resultado, r)
	}

	if err := linhas.Err(); err != nil {
		return []Reserva{}
	}

	return resultado
}
