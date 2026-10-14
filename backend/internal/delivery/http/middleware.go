package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupMiddlewares configures standard middleware stack on Echo instance.
func SetupMiddlewares(e *echo.Echo) {
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","status":${status},` +
			`"error":"${error}","latency_human":"${latency_human}"}` + "\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID},
		AllowCredentials: true,
	}))

	e.HTTPErrorHandler = CustomHTTPErrorHandler
}

// CustomHTTPErrorHandler formats uncaught errors into clean JSON error responses.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := "Internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		}
	}

	_ = c.JSON(code, map[string]interface{}{
		"code":    http.StatusText(code),
		"message": message,
	})
}

