package user

import "github.com/gofiber/fiber/v2"

func (r *userRouter) Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	return c.SendString("logged out")
}
