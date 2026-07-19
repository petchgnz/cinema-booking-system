package middleware

import (
	"log"
	"net/http"
	"strings"

	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies Firebase JWT and upserts the user into MongoDB.
// It injects user_id, firebase_uid, and role into the Gin context.
func AuthMiddleware(firebaseAuth *auth.Client, userRepo repository.UserRepository, adminEmail string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "missing authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		decoded, err := firebaseAuth.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			log.Printf("[Auth] Token verify failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "invalid token"})
			c.Abort()
			return
		}

		user := &model.User{
			FirebaseUID: decoded.UID,
			Email:       getStringClaim(decoded, "email"),
			Name:        getStringClaim(decoded, "name"),
			PhotoURL:    getStringClaim(decoded, "picture"),
		}

		savedUser, err := userRepo.Upsert(c.Request.Context(), user, adminEmail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to process user"})
			c.Abort()
			return
		}

		c.Set("user_id", savedUser.ID.Hex())
		c.Set("firebase_uid", decoded.UID)
		c.Set("role", string(savedUser.Role))

		c.Next()
	}
}

// helpers
func getStringClaim(token *auth.Token, key string) string {
	if val, ok := token.Claims[key].(string); ok {
		return val
	}
	return ""
}
