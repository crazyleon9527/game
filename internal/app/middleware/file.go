package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FileUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// File upload
		log.Println("FileUploadMiddleware") //用于等待文件上传完
		_, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Failed to upload file"})
			c.Abort()
			return
		}
		// Process the uploaded file
		// ...
		c.Next()
	}
}
