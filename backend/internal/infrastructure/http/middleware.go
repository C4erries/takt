package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	UserIDContextKey      contextKey = "userID"
	WorkspaceIDContextKey contextKey = "workspaceID"

	AuthCookieName = "token"
)

type JWTClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: 7 * 24 * time.Hour,
	}
}

func (m *JWTManager) GenerateToken(userID, workspaceID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := JWTClaims{
		UserID:      userID,
		WorkspaceID: workspaceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) ParseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (h *BcryptHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Skipper defines a function to skip middleware.
type Skipper func(c echo.Context) bool

func AuthMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return AuthMiddlewareWithSkipper(jwtManager, nil)
}

func AuthMiddlewareWithSkipper(jwtManager *JWTManager, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper != nil && skipper(c) {
				return next(c)
			}

			var tokenStr string

			// 1. Check HTTP-only cookie
			if cookie, err := c.Cookie(AuthCookieName); err == nil && cookie.Value != "" {
				tokenStr = cookie.Value
			}

			// 2. Fallback to Authorization Header
			if tokenStr == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenStr == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication credentials",
				})
			}

			claims, err := jwtManager.ParseToken(tokenStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				})
			}

			// Set in Echo context
			c.Set(string(UserIDContextKey), claims.UserID)
			c.Set(string(WorkspaceIDContextKey), claims.WorkspaceID)

			// Also set in standard context.Context
			reqCtx := context.WithValue(c.Request().Context(), UserIDContextKey, claims.UserID)
			reqCtx = context.WithValue(reqCtx, "userID", claims.UserID)
			reqCtx = context.WithValue(reqCtx, WorkspaceIDContextKey, claims.WorkspaceID)
			c.SetRequest(c.Request().WithContext(reqCtx))

			return next(c)
		}
	}
}

func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	val := c.Get(string(UserIDContextKey))
	if id, ok := val.(uuid.UUID); ok && id != uuid.Nil {
		return id, nil
	}
	return uuid.Nil, errors.New("unauthorized: user id not found in context")
}

func GetWorkspaceIDFromContext(c echo.Context) (uuid.UUID, error) {
	val := c.Get(string(WorkspaceIDContextKey))
	if id, ok := val.(uuid.UUID); ok && id != uuid.Nil {
		return id, nil
	}
	return uuid.Nil, errors.New("unauthorized: workspace id not found in context")
}

func SetAuthCookie(c echo.Context, token string, isProduction bool) {
	cookie := new(http.Cookie)
	cookie.Name = AuthCookieName
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	if isProduction {
		cookie.Secure = true
	}
	c.SetCookie(cookie)
}

func ClearAuthCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = AuthCookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(0, 0)
	c.SetCookie(cookie)
}

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

	if code >= http.StatusInternalServerError {
		log.Printf("[ERROR 500] %s %s: %+v", c.Request().Method, c.Request().RequestURI, err)
	}

	_ = c.JSON(code, map[string]interface{}{
		"code":    http.StatusText(code),
		"message": message,
	})
}
