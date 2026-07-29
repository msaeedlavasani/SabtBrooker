package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/auth"
)

type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	MobileKey  contextKey = "mobile"
)

// AuthRequired validates the JWT token and injects user info into context
func AuthRequired(jwtManager *auth.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "توکن احراز هویت ارائه نشده است",
				})
			}

			// Extract Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "فرمت توکن نامعتبر است",
				})
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				slog.Warn("token validation failed", "error", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "توکن نامعتبر یا منقضی شده است",
				})
			}

			// Inject user info into context
			c.Set("user_id", claims.UserID)
			c.Set("user_role", claims.Role)
			c.Set("mobile", claims.Mobile)

			return next(c)
		}
	}
}

// RequireRole restricts access to specific roles
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("user_role")
			if role == nil {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "دسترسی غیرمجاز",
				})
			}

			userRole := role.(string)
			for _, r := range roles {
				if userRole == r {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "شما مجوز دسترسی به این بخش را ندارید",
			})
		}
	}
}

// GetUserID extracts user ID from Echo context
func GetUserID(c echo.Context) string {
	id := c.Get("user_id")
	if id == nil {
		return ""
	}
	return id.(string)
}

// GetUserRole extracts user role from Echo context
func GetUserRole(c echo.Context) string {
	role := c.Get("user_role")
	if role == nil {
		return ""
	}
	return role.(string)
}

// CORS returns CORS middleware for development
func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Response().Header().Set("Access-Control-Max-Age", "86400")

			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

// RequestLogger logs each request with method, path, status, and latency
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			slog.Info("request",
				"method", req.Method,
				"path", req.URL.Path,
				"ip", c.RealIP(),
			)
			return next(c)
		}
	}
}

// Recover middleware catches panics
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("panic recovered",
						"error", r,
						"path", c.Request().URL.Path,
					)
					c.JSON(http.StatusInternalServerError, map[string]string{
						"error": "خطای داخلی سرور",
					})
				}
			}()
			return next(c)
		}
	}
}
