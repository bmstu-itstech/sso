package http_server

import (
	"context"
	"net/http"
	"strconv"

	_ "github.com/bmstu-itstech/sso/docs"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Auth interface {
	Login(ctx context.Context, appId int32, login string, password string) (token string, err error)
	RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error)

	UpdatePassword(ctx context.Context, id int64, newPassword string) error
	DeleteUser(ctx context.Context, userId int64) error

	SignIn(ctx context.Context, token string, appId int32) (tokenModel models.TokenInfo, err error)
	IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)

	UserInfo(ctx context.Context, id int64) (user models.UserServices, err error)
	UsersAll(ctx context.Context) ([]models.UserServices, error)

	UpdateTokenApp(ctx context.Context, token string, appId int32) (string, error)
}

const ssoAppID int32 = 0 // ID приложения

type ServerGin struct {
	auth     Auth
	cfg      *config.Config
	validate *validator.Validate
}

// Конструктор сервера
func New(auth Auth, cfg *config.Config) *ServerGin {
	return &ServerGin{
		auth:     auth,
		cfg:      cfg,
		validate: validator.New(),
	}
}
func (s *ServerGin) InitRouter() *gin.Engine {

	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	v1 := api.Group("/v1")
	{
		v1.GET("/ping", s.ping)
		v1.POST("/login", s.login)
		v1.POST("/register", s.register)

		userGroup := v1.Group("/user")
		userGroup.Use(s.authMiddleware())
		{
			userGroup.GET("/is_admin/:id", s.isAdmin)
			userGroup.GET("/info/:id", s.userInfo)
			userGroup.GET("/info", s.users)
			userGroup.POST("/update_token", s.updateToken)
			userGroup.PUT("/", s.updatePassword)
			userGroup.DELETE("/", s.removeUser)
		}
	}

	return router
}

// Helper для декодирования и валидации JSON
func (s *ServerGin) decodeAndValidate(c *gin.Context, v interface{}) error {
	if err := c.ShouldBindJSON(v); err != nil {
		return err
	}
	return s.validate.Struct(v)
}

// Обработчики маршрутов

// ping godoc
// @Summary     Health check
// @Description Проверка доступности сервиса.
// @Tags        Health
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /ping [get]
func (s *ServerGin) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// login godoc
// @Summary     Login
// @Description Логин пользователя. Возвращает JWT.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body models.LoginRequest true "Данные входа"
// @Success     200 {object} models.LoginResponse
// @Failure     400 {object} ErrorResponse "Некорректный JSON или не пройдена валидация"
// @Failure     401 {object} ErrorResponse "Неверные учётные данные"
// @Router      /login [post]
func (s *ServerGin) login(c *gin.Context) {
	var req models.LoginRequest
	if err := s.decodeAndValidate(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := s.auth.Login(c.Request.Context(), req.AppId, req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token})
}

// register godoc
// @Summary     Register
// @Description Регистрация нового пользователя.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body models.RegisterRequest true "Данные регистрации"
// @Success     200 {object} models.RegisterResponse
// @Failure     400 {object} ErrorResponse "Некорректный JSON или не пройдена валидация"
// @Failure     409 {object} ErrorResponse "Пользователь уже существует"
// @Router      /register [post]
func (s *ServerGin) register(c *gin.Context) {
	var req models.RegisterRequest
	if err := s.decodeAndValidate(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := s.auth.RegisterNewUser(c.Request.Context(), req.Login, req.Password, req.Email, req.FullName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusOK, models.RegisterResponse{UserId: userId})
}

// isAdmin godoc
// @Summary     Check admin
// @Description Проверка, является ли пользователь админом.
// @Tags        User
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "User ID"
// @Success     200 {object} models.IsAdminResponse
// @Failure     400 {object} ErrorResponse "Неверный формат id"
// @Failure     401 {object} ErrorResponse "Нет/невалидный Bearer token"
// @Failure     403 {object} ErrorResponse "Недостаточно прав"
// @Failure     404 {object} ErrorResponse "Пользователь не найден"
// @Router      /user/is_admin/{id} [get]
func (s *ServerGin) isAdmin(c *gin.Context) {
	var req models.IsAdminRequest
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	req.UserId = id

	userId := c.GetInt64("userId")
	isAdmin := c.GetBool("isAdmin")
	if !isAdmin && userId != req.UserId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	isAdmin, err = s.auth.IsAdmin(c.Request.Context(), req.UserId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, models.IsAdminResponse{IsAdmin: isAdmin})
}

// userInfo godoc
// @Summary     User info
// @Description Получение информации о пользователе по ID.
// @Tags        User
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "User ID"
// @Success     200 {object} models.User
// @Failure     400 {object} ErrorResponse "Неверный формат id"
// @Failure     401 {object} ErrorResponse "Нет/невалидный Bearer token"
// @Failure     403 {object} ErrorResponse "Недостаточно прав"
// @Failure     404 {object} ErrorResponse "Пользователь не найден"
// @Router      /user/info/{id} [get]
func (s *ServerGin) userInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	userId := c.GetInt64("userId")
	isAdmin, err := s.auth.IsAdmin(c.Request.Context(), userId)
	if err != nil || (!isAdmin && userId != id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userInfo, err := s.auth.UserInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// users godoc
// @Summary     Users list
// @Description Список всех пользователей (только для admin).
// @Tags        Admin
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} models.Users
// @Failure     401 {object} ErrorResponse "Нет/невалидный Bearer token"
// @Failure     403 {object} ErrorResponse "Требуются права admin"
// @Failure     500 {object} ErrorResponse "Ошибка получения списка"
// @Router      /user/info [get]
func (s *ServerGin) users(c *gin.Context) {
	userId := c.GetInt64("userId")
	isAdmin, err := s.auth.IsAdmin(c.Request.Context(), userId)
	if err != nil || !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	users, err := s.auth.UsersAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	usersResp := make([]models.User, len(users))
	for i := range users {
		usersResp[i] = models.User{
			UserId:    users[i].ID,
			Login:     users[i].Login,
			Email:     users[i].Email,
			FullName:  users[i].FullName,
			IsAdmin:   users[i].IsAdmin,
			CreatedAt: users[i].CreatedAt,
			UpdatedAt: users[i].UpdatedAt,
		}
	}
	clear(users)
	c.JSON(http.StatusOK, models.Users{Users: usersResp})
}

// updateToken godoc
// @Summary     Update app token
// @Description Обновляет токен приложения (refresh). Берёт текущий JWT из Authorization.
// @Tags        Auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body models.UpdateTokenRequest true "ID приложения"
// @Success     200 {object} models.UpdateTokenResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /user/update_token [post]
func (s *ServerGin) updateToken(c *gin.Context) {
	var req models.UpdateTokenRequest
	if err := s.decodeAndValidate(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	oldToken := c.GetString("jwtCleanToken")
	newToken, err := s.auth.UpdateTokenApp(c.Request.Context(), oldToken, req.AppId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update newToken"})
		return
	}

	c.JSON(http.StatusOK, models.UpdateTokenResponse{Token: newToken})
}

// updatePassword godoc
// @Summary     Update password
// @Description Обновление пароля (сам себе или admin).
// @Tags        User
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body models.UpdatePasswordRequest true "Новый пароль"
// @Success     200 {object} models.UpdatePasswordResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /user/ [put]
func (s *ServerGin) updatePassword(c *gin.Context) {
	var req models.UpdatePasswordRequest
	if err := s.decodeAndValidate(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userId")
	isAdmin, err := s.auth.IsAdmin(c.Request.Context(), userId)
	if err != nil || (!isAdmin && userId != req.UserId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if err := s.auth.UpdatePassword(c.Request.Context(), req.UserId, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, models.UpdatePasswordResponse{Message: "Password updated"})
}

// removeUser godoc
// @Summary     Remove user
// @Description Удаление пользователя (сам себя или admin).
// @Tags        User
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body models.RemoveUserRequest true "User ID"
// @Success     200 {object} models.RemoveUserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /user/ [delete]
func (s *ServerGin) removeUser(c *gin.Context) {
	var req models.RemoveUserRequest
	if err := s.decodeAndValidate(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetInt64("userId")
	isAdmin, err := s.auth.IsAdmin(c.Request.Context(), userId)
	if err != nil || (!isAdmin && userId != req.UserId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if err := s.auth.DeleteUser(c.Request.Context(), req.UserId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, models.RemoveUserResponse{Message: "User deleted"})
}
