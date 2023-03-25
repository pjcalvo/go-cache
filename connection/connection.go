package connection

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName = "responses.data"
)

type Response struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type SqlLiteClient struct {
	db *sql.DB
}

func Init() (*SqlLiteClient, error) {

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	query := `
    CREATE TABLE IF NOT EXISTS links 
	(id INTEGER PRIMARY KEY, link_id TEXT, value TEXT);
    `

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &SqlLiteClient{
		db: db,
	}, nil
}

func (s SqlLiteClient) Get(id string) (*Response, error) {
	row := s.db.QueryRow("SELECT link_id, value FROM links where link_id = ?", id)

	var res Response
	if err := row.Scan(&res.Id, &res.Value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &res, nil
}

func (s SqlLiteClient) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM links WHERE link_id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return err
}

func (s SqlLiteClient) Add(r Response) error {
	_, err := s.db.Exec("INSERT INTO links(value, link_id) values(?,?)", r.Value, r.Id)
	if err != nil {
		return err
	}

	return nil
}
