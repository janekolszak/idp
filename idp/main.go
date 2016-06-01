package main

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Consent struct {
	Challenge string `form:"challenge" binding:"required"`
}

func handleConsent(c *gin.Context) {
	var data Consent
	if c.Bind(&data) == nil {
	}

	c.JSON(200, gin.H{"challenge": data.Challenge})
}

func main() {
	r := gin.Default()
	r.GET("/", handleConsent)

	r.Run(":3000")
}
