package middleware

import (
	"net/http"

	"cinema-booking/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware must be used after AuthMiddleware.
// It checks that the authenticated user has the "admin" role.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != string(model.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
