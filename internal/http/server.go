package http_server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	api := router.Group("/api")
	v1 := api.Group("/v1")
	{
		v1.GET("/ping", s.ping)
		v1.POST("/login", s.login)
		v1.POST("/register", s.register)

		userGroup := v1.Group("/user", s.authMiddleware())
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
func (s *ServerGin) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

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
