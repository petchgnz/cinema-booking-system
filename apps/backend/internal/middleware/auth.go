package middleware

import (
	"net/http"
	"strings"

	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// auth middleware: verify firebaseID token & inject user เข้า context
func AuthMiddleware(firebaseAuth *auth.Client, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get token from header (authorization)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "missing authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// verify
		decoded, err := firebaseAuth.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
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

		savedUser, err := userRepo.Upsert(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to process user"})
			c.Abort()
			return
		}

		c.Set("user_id", savedUser.ID.Hex())
		c.Set("firebase_uid", decoded.UID)

		// continue to handler (controller)
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
