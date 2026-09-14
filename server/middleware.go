package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	// maxRequestBodySize limits JSON bodies of API requests.
	maxRequestBodySize = 1 << 20
	// csrfHeader must be sent with requests that change data. A cross-site form or
	// a simple request can't set it, and custom headers from other origins need
	// a CORS preflight, which the API doesn't allow.
	csrfHeader      = "X-Requested-With"
	csrfHeaderValue = "fetch"
)

func limitRequestBody(c *gin.Context) {
	if c.Request.Body != nil {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)
	}
	c.Next()
}

func requireCSRFHeader(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		if c.GetHeader(csrfHeader) != csrfHeaderValue {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Запрос отклонён"})
			return
		}
	}
	c.Next()
}
