package http_response

import "github.com/gofiber/fiber/v2"

type OkResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func HttpOkResponse(c *fiber.Ctx, code string, message string) error {
	return c.Status(fiber.StatusOK).JSON(OkResponse{Code: code, Message: message})
}
