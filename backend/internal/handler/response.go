package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// OK sends a 200 JSON response
func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// Created sends a 201 JSON response
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, data)
}

// NoContent sends a 204 response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// BadRequest sends a 400 JSON error
func BadRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": msg})
}

// Unauthorized sends a 401 JSON error
func Unauthorized(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": msg})
}

// Forbidden sends a 403 JSON error
func Forbidden(c echo.Context, msg string) error {
	return c.JSON(http.StatusForbidden, map[string]string{"error": msg})
}

// NotFound sends a 404 JSON error
func NotFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, map[string]string{"error": msg})
}

// Conflict sends a 409 JSON error
func Conflict(c echo.Context, msg string) error {
	return c.JSON(http.StatusConflict, map[string]string{"error": msg})
}

// Unprocessable sends a 422 JSON error with details
func Unprocessable(c echo.Context, msg string, details interface{}) error {
	return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
		"error":   msg,
		"details": details,
	})
}

// InternalError sends a 500 JSON error
func InternalError(c echo.Context, msg string) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": msg})
}

// PageParams extracts pagination params from query
type PageParams struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	SortBy   string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (p *PageParams) Defaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}
