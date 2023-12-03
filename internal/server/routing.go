package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"accelerator/internal/session"
	"accelerator/models"
	"github.com/gofiber/fiber/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) SetupRouting() {
	s.conn.Get("/hemlo", func(c *fiber.Ctx) error {
		return c.SendString("hemlo!")
	})
	s.conn.Get("/api/brands/open", func(c *fiber.Ctx) error { // api/brands/open?offset=x&limit=y
		offset, _ := strconv.Atoi(c.Params("offset", "0"))
		limit, _ := strconv.Atoi(c.Params("limit", "30"))
		brands, err := s.db.GetOpenBrands()
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		var tosend []models.Brand
		for i := range brands {
			if i >= offset && i < limit+offset { // in [offset; limit + offset)
				tosend = append(tosend, brands[i])
			}
		}
		marshal, err := json.Marshal(tosend)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		return c.Send(marshal)
	})
	s.conn.Post("/api/users/register", func(c *fiber.Ctx) error {
		var req models.Person
		err := json.Unmarshal(c.Body(), &req)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusBadRequest)
		}
		owner := models.NewOwner(&req)
		err = s.db.InsertOwner(owner)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		return c.SendStatus(http.StatusOK)
	})
	s.conn.Post("/api/users/login", func(c *fiber.Ctx) error {
		var req models.LoginData
		err := json.Unmarshal(c.Body(), &req)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusBadRequest)
		}
		actualpassword, err := s.db.GetPasswordByEmail(req.Email)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		isGood := bcrypt.CompareHashAndPassword([]byte(actualpassword), []byte(req.Password))
		if isGood != nil {
			return c.SendStatus(http.StatusForbidden)
		}
		newsession := session.NewSession(time.Now().Add(s.sessionLen))
		// apparently tarantool hates "local" timezones, so we'll just make everything the same - no timezone
		newsession.ExpTime = newsession.ExpTime.In(time.FixedZone(datetime.NoTimezone, 0))
		err = s.scash.StoreSession(&newsession)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		marshal, err := json.Marshal(newsession)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		return c.Send(marshal)
	})
	s.conn.Get("/api/users/checklogin", func(c *fiber.Ctx) error {
		var req session.JustToken
		err := json.Unmarshal(c.Body(), &req)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusBadRequest)
		}
		ses, err := s.scash.FindSession(req.Token)
		if err != nil {
			s.log.Errorln(err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		empty := session.Session{}
		if ses == empty || ses.IsExpired() {
			err = s.scash.DeleteSession(req.Token)
			if err != nil {
				s.log.Errorln(err)
			}
			return c.SendString("false")
		}
		return c.SendString("true")
	})
}
