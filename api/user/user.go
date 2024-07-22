package user

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/zura-t/bookstore_fiber/api"
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
	err := db.Find(&users).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

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
		err := fmt.Errorf("Can't get payload")
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	var user models.User
	db := database.DbConn

	err := db.Find(&user, data.UserId).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse(err))
	}
	res := ConvertUser(user)
	return c.JSON(res)
}

func GetUser(c *fiber.Ctx) error {
	var req UserId
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	var user models.User
	db := database.DbConn

	err := db.Find(&user, req.Id).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse(err))
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
	IsAuthor  bool      `json:"is_admin"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

func Register(c *fiber.Ctx) error {
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error(err)
		return c.Status(503).JSON(api.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	db := database.DbConn

	var user models.User
	err := db.Find(&user, models.User{Email: req.Email}).Error
	if user != (models.User{}) {
		err = fmt.Errorf("User with this email already exists")
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	new_user := models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	err = db.Create(&new_user).Error
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	res := ConvertUser(new_user)

	return c.JSON(res)
}

func ConvertUser(user models.User) UserResponse {
	return UserResponse{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAuthor:  user.IsAuthor,
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
		return c.Status(503).JSON(api.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	db := database.DbConn

	var user models.User
	err := db.Find(&user, models.User{Email: req.Email}).Error
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse(err))
	}

	err = pkg.CheckPassword(req.Password, user.Password)
	if err != nil {
		err = fmt.Errorf("Error incorrect password, %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	accessToken, accessPayload, err := token.Jwtmaker.CreateToken(user.ID, user.Email, config.Cfg.AccessTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	refreshToken, refreshPayload, err := token.Jwtmaker.CreateToken(user.ID, user.Email, config.Cfg.RefreshTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
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
		err := fmt.Errorf("Can't get payload")
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	var user UserUpdate
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	db := database.DbConn
	err := db.Model(&models.User{}).Where("id = ?", data.UserId).Updates(&user).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	var res models.User
	err = db.Find(&res, data.UserId).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}
	if res == (models.User{}) {
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse(err))
	}

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
		err := fmt.Errorf("can't renew the token")
		return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse(err))
	}

	refreshPayload, err := token.Jwtmaker.VerifyToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse(err))
	}

	accessToken, accessPayload, err := token.Jwtmaker.CreateToken(refreshPayload.UserId, refreshPayload.Email, config.Cfg.AccessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
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
	return c.SendString("logged out")
}

func DeleteMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}

	db := database.DbConn
	var user models.User
	err := db.Delete(&user, data.UserId).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	return c.SendString("Profile deleted")
}

func BecomeAuthor(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse(err))
	}
	db := database.DbConn
	err := db.Model(&models.User{}).Where("id = ?", data.UserId).Update("IsAuthor", true).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse(err))
	}

	return c.SendString("you became an author")
}
