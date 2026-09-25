package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PgConnection struct {
	url string
}

func NewPgConnection(url string) *PgConnection {
	return &PgConnection{
		url: url,
	}
}

func (c *PgConnection) Connect(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, c.url)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to postgres database: error=%s", err.Error())
	}

	return conn, nil
}

func (c *PgConnection) Close(ctx context.Context, conn *pgx.Conn) error {
	err := conn.Close(ctx)
	if err != nil {
		return fmt.Errorf("fail to close connection to postgres database: error=%s", err.Error())
	}

	return nil
}
