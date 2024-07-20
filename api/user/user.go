package user

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/database"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserId struct {
	Id uint `uri:"id" validate:"required,min=1"`
}

func GetUsers(c *fiber.Ctx) error {
	var users []models.User
	db := database.DbConn
	db.Find(&users)
	res := make([]*UserResponse, len(users))
	for index, v := range users {
		user := ConvertUser(v)
		res[index] = &user
	}
	return c.JSON(res)
}

func GetMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Can't get payload")
	}

	var user models.User
	db := database.DbConn

	db.Find(&user, data.UserId)
	if user == (models.User{}) {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	res := ConvertUser(user)
	return c.JSON(res)
}

func GetUser(c *fiber.Ctx) error {
	var req UserId
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	var user models.User
	db := database.DbConn

	db.Find(&user, req.Id)
	if user == (models.User{}) {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	res := ConvertUser(user)
	return c.JSON(res)
}

type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	Id        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

func Register(c *fiber.Ctx) error {
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error(err)
		return c.Status(503).SendString(err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	db := database.DbConn

	var user models.User
	err := db.Find(&user, models.User{Email: req.Email}).Error
	if user != (models.User{}) {
		return c.Status(fiber.StatusBadRequest).SendString("User with this email already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	new_user := models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	err = db.Create(&new_user).Error
	if err != nil {
		err = fmt.Errorf("failed to create user: %s", err)
		return c.Status(400).SendString(err.Error())
	}

	res := ConvertUser(new_user)

	return c.JSON(res)
}

func ConvertUser(user models.User) UserResponse {
	return UserResponse{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponse struct {
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  UserResponse `json:"user"`
}

func Login(c *fiber.Ctx) error {
	var req LoginUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error(err)
		return c.Status(503).SendString(err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	db := database.DbConn

	var user models.User
	err := db.Find(&user, models.User{Email: req.Email}).Error
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusBadRequest).SendString("User not found")
	}

	err = pkg.CheckPassword(req.Password, user.Password)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusBadRequest).SendString("Error incorrect password")
	}

	accessToken, accessPayload, err := token.Jwtmaker.CreateToken(user.ID, user.Email, config.Cfg.AccessTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	refreshToken, refreshPayload, err := token.Jwtmaker.CreateToken(user.ID, user.Email, config.Cfg.RefreshTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	res := LoginUserResponse{
		User:                  ConvertUser(user),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}

	return c.JSON(res)
}

type UserUpdate struct {
	Name string `json:"name" validate:"required"`
}

func UpdateMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Can't get payload")
	}

	var user UserUpdate
	if err := c.BodyParser(&user); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	db := database.DbConn
	db.Model(&models.User{}).Where("id = ?", data.UserId).Updates(&user)
	var res models.User
	db.Find(&res, data.UserId)
	resp := ConvertUser(res)
	return c.JSON(resp)
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func RenewAccessToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("can't renew the token")
	}

	refreshPayload, err := token.Jwtmaker.VerifyToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	accessToken, accessPayload, err := token.Jwtmaker.CreateToken(refreshPayload.UserId, refreshPayload.Email, config.Cfg.AccessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("can't create new token")
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	return c.JSON(rsp)
}

func Logout(c *fiber.Ctx) error {
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
	return c.Status(fiber.StatusOK).SendString("logged out")
}

func DeleteMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Can't get payload")
	}

	db := database.DbConn
	var user models.User
	db.Delete(&user, data.UserId)
	return c.SendString("Profile deleted")
}
