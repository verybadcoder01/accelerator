package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"accelerator/models"
	"github.com/gofiber/fiber/v2"
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
}
