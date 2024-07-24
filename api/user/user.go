package user

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type userRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
	token  *token.JwtMaker
}

func NewuserRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &userRouter{log, config, db, token}

	app.Post("/register", r.Register)
	app.Post("/login", r.Login)
	app.Post("/renew_token", r.RenewAccessToken)
	app.Post("/logout", r.Logout)

	authRoutes := app.Group("/", auth.New(log, token))

	authRoutes.Get("/users", r.GetUsers)
	authRoutes.Get("/users/my_profile", r.GetMyProfile)
	authRoutes.Get("/users/:id", r.GetUser)
	authRoutes.Patch("/users/my_profile", r.UpdateMyProfile)
	authRoutes.Delete("/users/my_profile", r.DeleteMyProfile)
	authRoutes.Patch("/users/author", r.BecomeAuthor)
}

type UserId struct {
	Id uint `uri:"id" validate:"required,min=1"`
}

func (r *userRouter) GetUsers(c *fiber.Ctx) error {
	var users []models.User
	err := r.db.Find(&users).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*UserResponse, len(users))
	for index, v := range users {
		user := ConvertUser(v)
		res[index] = &user
	}
	return c.JSON(res)
}

func (r *userRouter) GetMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User

	err := r.db.Find(&user, data.UserId).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
	}
	res := ConvertUser(user)
	return c.JSON(res)
}

func (r *userRouter) GetUser(c *fiber.Ctx) error {
	var req UserId
	if err := c.ParamsParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User

	err := r.db.Find(&user, req.Id).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
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
	IsAuthor  bool      `json:"is_author"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

func (r *userRouter) Register(c *fiber.Ctx) error {
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(503).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User
	err := r.db.Find(&user, models.User{Email: req.Email}).Error
	if user != (models.User{}) {
		err = fmt.Errorf("User with this email already exists")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	new_user := models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	err = r.db.Create(&new_user).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
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

func (r *userRouter) Login(c *fiber.Ctx) error {
	var req LoginUserRequest
	if err := c.BodyParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(503).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User
	err := r.db.Find(&user, models.User{Email: req.Email}).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	if user == (models.User{}) {
		err = fmt.Errorf("User not found")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
	}

	err = pkg.CheckPassword(req.Password, user.Password)
	if err != nil {
		err = fmt.Errorf("Error incorrect password, %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	accessToken, accessPayload, err := r.token.CreateToken(user.ID, user.Email, r.config.AccessTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	refreshToken, refreshPayload, err := r.token.CreateToken(user.ID, user.Email, r.config.RefreshTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
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

func (r *userRouter) UpdateMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user UserUpdate
	if err := c.BodyParser(&user); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	err := r.db.Model(&models.User{}).Where("id = ?", data.UserId).Updates(&user).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	var res models.User
	err = r.db.Find(&res, data.UserId).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}
	if res == (models.User{}) {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
	}

	resp := ConvertUser(res)
	return c.JSON(resp)
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (r *userRouter) RenewAccessToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		err := fmt.Errorf("can't renew the token")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusUnauthorized).JSON(pkg.ErrorResponse(err))
	}

	refreshPayload, err := r.token.VerifyToken(refreshToken)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusUnauthorized).JSON(pkg.ErrorResponse(err))
	}

	accessToken, accessPayload, err := r.token.CreateToken(refreshPayload.UserId, refreshPayload.Email, r.config.AccessTokenDuration)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	return c.JSON(rsp)
}

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

func (r *userRouter) DeleteMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User
	err := r.db.Delete(&user, data.UserId).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.SendString("Profile deleted")
}

func (r *userRouter) BecomeAuthor(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}
	err := r.db.Model(&models.User{}).Where("id = ?", data.UserId).Update("IsAuthor", true).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.SendString("you became an author")
}
