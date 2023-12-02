package server

import (
	"time"

	"accelerator/internal/db"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	conn       *fiber.App
	log        *log.Logger
	sessionLen time.Duration
	db         *db.Database
}

func NewServer(db *db.Database, log *log.Logger, slensec int64) Server {
	return Server{conn: fiber.New(), log: log, sessionLen: time.Duration(slensec) * time.Second, db: db}
}

func (s *Server) SetupRouting() {
	setupRouting(s.conn)
}

func (s *Server) ListenAndServe(port string) error {
	err := s.conn.Listen(port)
	return err
}
