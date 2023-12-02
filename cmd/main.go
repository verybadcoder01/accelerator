package main

import (
	db "accelerator/internal/db"
	"accelerator/internal/server"
	"accelerator/internal/sessioncash"
	"accelerator/logging"
	log "github.com/sirupsen/logrus"
	"github.com/tarantool/go-tarantool/v2"

	"accelerator/config"
)

func main() {
	conf, err := config.ParseConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.SetupLogging(conf.LogPath, conf.LogLevel)
	logger.Infoln("Here we go")
	dconn := db.NewDb(conf.DbPath, logger)
	err = dconn.CreateTables()
	if err != nil {
		log.Fatal(err)
	}
	cashDb, err := sessioncash.NewTarantoolCashDB(conf.SessionCashPath, conf.CashUser, conf.CashPassword, "sessions", tarantool.Opts{
		Reconnect: 10, MaxReconnects: 3,
	})
	if err != nil {
		log.Fatal(err)
	}
	s := server.NewServer(&dconn, &cashDb, logger, conf.SessionLenSec)
	s.SetupRouting()
	err = s.ListenAndServe(conf.Port)
	if err != nil {
		log.Errorln(err)
	}
}
