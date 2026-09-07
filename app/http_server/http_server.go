package http_server

import (
	"github.com/gofiber/fiber/v2"
)

// For Fiber middlewares
func (r *Route) Use(args ...interface{}) {
	r.fiber.Use(args...)
}

// For Fiber route grouping
func (r *Route) Group(prefix string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Group(prefix, handlers...)
}

// GET register service endpoint for HTTP GET
func (r *Route) Get(path string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Get(path, handlers...)
}

// POST register service endpoint for HTTP POST
func (r *Route) Post(path string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Post(path, handlers...)
}

// PUT register service endpoint for HTTP PUT
func (r *Route) Put(path string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Put(path, handlers...)
}

// PATCH register service endpoint for HTTP PATCH
func (r *Route) Patch(path string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Patch(path, handlers...)
}

// DELETE register service endpoint for HTTP DELETE
func (r *Route) Delete(path string, handlers ...func(*fiber.Ctx) error) fiber.Router {
	return r.fiber.Delete(path, handlers...)
}
