package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/XaviFP/notifications/publisher/internal/publisher"
	"github.com/gorilla/websocket"
)

var broker publisher.Broker
var addr = flag.String("addr", "0.0.0.0:8081", "http service address")
var dbHost = flag.String("dbHost", "db", "Database host")
var dbPort = flag.String("dbPort", "5432", "Database port")
var dbUser = flag.String("dbUser", "y", "Database user")
var dbPassword = flag.String("dbPassword", "y", "Database password")
var dbName = flag.String("dbName", "y", "Database name")
var dbSSLMode = flag.String("dbSSLMode", "disable", "SSL mode for DB connection")
var initialCredits = flag.Int("initialCredits", 10, "Initial credits for each user")
var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()
	dbConfig := publisher.DbConfig{
		Host:     *dbHost,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPassword,
		Name:     *dbName,
		SSLMode:  *dbSSLMode,
	}
	broker = publisher.NewBroker(dbConfig, *initialCredits)
	s := publisher.NewServer(broker)
	s.RegistersRoutes()
	s.Start()

	log.Fatal(http.ListenAndServe(*addr, nil))
}
