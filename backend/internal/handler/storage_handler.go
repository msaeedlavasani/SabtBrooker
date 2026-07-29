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
	store *storage.FileStorage
}

func NewStorageHandler(store *storage.FileStorage) *StorageHandler {
	return &StorageHandler{store: store}
}

func (h *StorageHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/presigned-url", h.GetPresignedURL)
}

func (h *StorageHandler) GetPresignedURL(c echo.Context) error {
	fileName := c.QueryParam("name")
	contentType := c.QueryParam("type") // e.g. application/pdf, image/jpeg

	// Validate file extension/type
	allowedTypes := map[string]bool{
		"application/pdf":                               true,
		"image/jpeg":                                    true,
		"image/png":                                     true,
		"application/octet-stream":                      true, // for DWG/DXF
		"application/x-autocad":                         true,
		"application/acad":                              true,
		"image/vnd.dwg":                                 true,
		"image/x-dwg":                                   true,
	}

	if contentType != "" && !allowedTypes[contentType] {
		return BadRequest(c, "نوع فایل مجاز نیست. فقط PDF، تصاویر و نقشه‌های فنی پذیرفته می‌شوند")
	}

	if fileName == "" {
		fileName = uuid.New().String()
	} else {
		// Append UUID to avoid collisions
		fileName = fmt.Sprintf("%s-%s", uuid.New().String(), fileName)
	}

	// Generate PUT URL for upload (valid for 15 minutes)
	url, err := h.store.PresignedUploadURL(context.Background(), fileName, 15*time.Minute)
	if err != nil {
		return InternalError(c, "خطا در ایجاد لینک بارگذاری")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"upload_url": url,
		"file_id":    fileName,
	})
}
