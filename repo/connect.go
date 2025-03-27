package repo

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Postgres struct {
	conn *sql.DB
}

func SetupDB(dsn string) (Postgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return Postgres{}, err
	}

	if err = db.Ping(); err != nil {
		return Postgres{}, err
	}

	return Postgres{conn: db}, nil
}

func (p *Postgres) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.conn.Exec(query, args...)
}

func (p *Postgres) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.conn.QueryRow(query, args...)
}

func (p *Postgres) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.conn.Query(query, args...)
}

func (p *Postgres) Conn() *sql.DB {
	return p.conn
}
