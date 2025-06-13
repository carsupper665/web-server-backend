package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login handles user login. For demo purposes it returns a fake token.
func Login(c *gin.Context) {
	// TODO: implement real authentication
	c.JSON(http.StatusOK, gin.H{"token": "demo-token"})
}
