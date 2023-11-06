package database

import (
	"context"
	"log"
	"os"
	"top_100_billboard_golang/environment"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbConn struct {
	Pool   *pgxpool.Pool
	Cancel context.CancelFunc
	Ctx    context.Context
}

var conn dbConn

func ConnectionSupabase() {
	cCtx, cCancel := context.WithCancel(context.Background())

	cPool, err := pgxpool.New(cCtx, os.Getenv("CONN_STRING_POSTGRES"))
	if err != nil {
		log.Fatal("cannot connect to database")
	}

	sqlText, err := os.ReadFile(environment.GetInitSQLPath())
	if err != nil {
		log.Fatal("not yet")
	}

	_, err = cPool.Exec(cCtx, string(sqlText))
	if err != nil {
		log.Fatal("pgPool not yet initiated")
	}

	conn = dbConn{
		Pool:   cPool,
		Cancel: cCancel,
		Ctx:    cCtx,
	}

	initWrappers()
}

func CloseConnection() {
	conn.Pool.Close()
	conn.Cancel()
}

func initWrappers() {
	SongInfoWrapper = songInfoWrapper{dbConn: &conn}
}
