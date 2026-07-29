package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/storage"
)

type StorageHandler struct {
	store *storage.MinIOStorage
}

func NewStorageHandler(store *storage.MinIOStorage) *StorageHandler {
	return &StorageHandler{store: store}
}

func (h *StorageHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/presigned-url", h.GetPresignedURL)
}

func (h *StorageHandler) GetPresignedURL(c echo.Context) error {
	fileName := c.QueryParam("name")
	if fileName == "" {
		fileName = uuid.New().String()
	} else {
		// Append UUID to avoid collisions
		fileName = fmt.Sprintf("%s-%s", uuid.New().String(), fileName)
	}

	// Generate PUT URL for upload (valid for 15 minutes)
	url, err := h.store.GeneratePresignedPutURL(context.Background(), fileName, 15*time.Minute)
	if err != nil {
		return InternalError(c, "خطا در ایجاد لینک بارگذاری")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"upload_url": url,
		"file_id":    fileName,
	})
}
