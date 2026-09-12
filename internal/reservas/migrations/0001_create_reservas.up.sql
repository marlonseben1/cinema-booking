CREATE TABLE IF NOT EXISTS reservas (
	id         SERIAL PRIMARY KEY,
	filme_id   TEXT NOT NULL,
	assento_id TEXT NOT NULL,
	usuario_id TEXT NOT NULL,
	status     TEXT NOT NULL,
	UNIQUE (filme_id, assento_id)
);
