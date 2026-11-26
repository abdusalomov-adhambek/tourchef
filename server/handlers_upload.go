package main

import (
	"crypto/rand"
	"encoding/hex"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const maxUploadBytes = 10 * 1024 * 1024 // 10 MB

var imageExtByMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func saveImage(c *fiber.Ctx, file *multipart.FileHeader) (string, error) {
	if file.Size > maxUploadBytes {
		return "", fiber.NewError(fiber.StatusBadRequest, "each image must be smaller than 10 MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if imageExtByMIME[mime.TypeByExtension(ext)] == "" {
		ext = imageExtByMIME[file.Header.Get("Content-Type")]
	}
	if ext == "" || imageExtByMIME[mime.TypeByExtension(ext)] == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "only jpg, png, webp and gif images are allowed")
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	name := hex.EncodeToString(buf) + ext
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return "", err
	}
	if err := c.SaveFile(file, filepath.Join(cfg.UploadDir, name)); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// handleUpload — bir yoki bir nechta rasm qabul qiladi.
// Maydon nomlari: multiimage, file yoki image.
// Qaytaradi: {url: birinchi, urls: [...], files: [...]}
func handleUpload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "multipart/form-data with an image file is required")
	}

	var headers []*multipart.FileHeader
	for _, field := range []string{"multiimage", "file", "image"} {
		headers = append(headers, form.File[field]...)
	}
	if len(headers) == 0 {
		return fail(c, fiber.StatusBadRequest, "image file is required (field: multiimage)")
	}

	urls := make([]string, 0, len(headers))
	files := make([]string, 0, len(headers))
	for _, h := range headers {
		url, err := saveImage(c, h)
		if err != nil {
			if fe, ok := err.(*fiber.Error); ok {
				return fail(c, fe.Code, fe.Message)
			}
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		urls = append(urls, url)
		files = append(files, h.Filename)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"url":   urls[0],
		"urls":  urls,
		"files": files,
	})
}
