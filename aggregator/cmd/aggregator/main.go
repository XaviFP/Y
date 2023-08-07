package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	aggregator "github.com/XaviFP/notifications/aggregator/internal"
	_ "github.com/lib/pq"
	"github.com/tilinna/clock"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var dbHost = flag.String("dbHost", "db", "Database host")
var dbPort = flag.String("dbPort", "5432", "Database port")
var dbUser = flag.String("dbUser", "y", "Database user")
var dbPassword = flag.String("dbPassword", "y", "Database password")
var dbName = flag.String("dbName", "y", "Database name")
var dbSSLMode = flag.String("dbSSLMode", "disable", "SSL mode for DB connection")

func main() {
	flag.Parse()
	dbConfig := aggregator.DbConfig{
		Host:     *dbHost,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPassword,
		Name:     *dbName,
		SSLMode:  *dbSSLMode,
	}

	db, err := sql.Open("postgres", dbConfig.String())
	if err != nil {
		panic(err)
	}
	articleRepo := aggregator.NewArticleRepository(db, clock.Realtime())

	s := aggregator.NewServer(articleRepo)
	s.RegistersRoutes()

	log.Fatal(http.ListenAndServe(*addr, nil))
}
