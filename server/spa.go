package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// registerSPA serves the frontend built into dir. Hashed assets are cached forever;
// any other GET request outside the API gets index.html, and the client router shows the page.
func registerSPA(router *gin.Engine, dir string) {
	assets := router.Group("/assets", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	})
	assets.Static("/", filepath.Join(dir, "assets"))

	// Files from the frontend's public folder, e.g. the favicon.
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name() != "index.html" {
				router.StaticFile("/"+entry.Name(), filepath.Join(dir, entry.Name()))
			}
		}
	}

	index := filepath.Join(dir, "index.html")
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Не найдено"})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}
		// index.html refers to the current asset hashes, so it must not be cached.
		c.Header("Cache-Control", "no-cache")
		c.File(index)
	})
}
