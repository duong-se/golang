package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func run() error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
