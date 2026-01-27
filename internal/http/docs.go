package http_server

// Этот файл нужен только для swaggo (swag init), чтобы:
// 1) задать метаданные API
// 2) описать схемы авторизации
// 3) держать общие модели ошибок/ответов
//
// Важно: эти комментарии читает swag и на их основе генерит Swagger 2.0.

// @title           SSO API
// @version         1.0.0
// @description     Сервис Single Sign-On (SSO): регистрация, логин, проверка JWT и управление пользователями.
// @description
// @description     Все приватные методы требуют заголовок: Authorization: Bearer <JWT>.
// @description
// @termsOfService  https://example.com/terms
//
// @contact.name    SSO team
// @contact.url     https://github.com/bmstu-itstech/sso
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization

// ErrorResponse — единый формат ошибок для фронтенда.
// В текущей реализации сервер возвращает {"error": "..."}.
// Мы фиксируем это как контракт.
//
// swagger:model ErrorResponse
// @description Ошибка запроса.
// @description Поле error — человекочитаемое описание.
// @description Пример: {"error":"Permission denied"}
type ErrorResponse struct {
	Error string `json:"error" example:"Permission denied"`
}
