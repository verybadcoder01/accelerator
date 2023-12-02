package main

import (
	db "accelerator/internal/db"
	"accelerator/internal/server"
	"accelerator/logging"
	log "github.com/sirupsen/logrus"

	"accelerator/config"
)

func main() {
	conf, err := config.ParseConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.SetupLogging(conf.LogPath, conf.LogLevel)
	dconn := db.NewDb(conf.DbPath, logger)
	err = dconn.CreateTables()
	if err != nil {
		log.Fatal(err)
	}
	s := server.NewServer(&dconn, logger, conf.SessionLenSec)
	s.SetupRouting()
	err = s.ListenAndServe(conf.Port)
	if err != nil {
		log.Errorln(err)
	}
}
