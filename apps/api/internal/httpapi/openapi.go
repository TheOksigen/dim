package httpapi

import (
	_ "embed"

	"github.com/gofiber/fiber/v3"
)

//go:embed openapi.json
var openAPIDocument []byte

func (a *App) openAPISpec(c fiber.Ctx) error {
	c.Type("json")
	return c.Send(openAPIDocument)
}
