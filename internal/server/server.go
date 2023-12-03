package server

import (
	"time"

	"accelerator/internal/db"
	"accelerator/internal/sessioncash"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	conn       *fiber.App
	log        *log.Logger
	sessionLen time.Duration
	db         *db.Database
	scash      sessioncash.CashDb
}

func NewServer(db *db.Database, cash sessioncash.CashDb, log *log.Logger, slensec int64) Server {
	return Server{conn: fiber.New(), log: log, sessionLen: time.Duration(slensec) * time.Second, db: db, scash: cash}
}

func (s *Server) ListenAndServe(port string) error {
	err := s.conn.Listen(port)
	return err
}
