package api

import (
	"bytes"
	"context"
	"fmt"
	"hood/internal/repository"
	"hood/internal/resolver"
	"hood/internal/util"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartApi(port int, r resolver.Resolver) error {
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	plaidRepository := repository.NewPlaidRepository(
		secrets.Plaid.ClientID,
		secrets.Plaid.Secret,
	)

	router := gin.Default()

	router.Use(blockBots)
	router.Use(cors.Default())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{"message": "welcome to hood"})
	})

	router.POST("/plaidLinkToken", func(ctx *gin.Context) {

		linkToken, err := plaidRepository.GetLinkToken(context.Background())
		if err != nil {
			returnErrorJson(err, ctx)
			return
		}

		ctx.JSON(200, map[string]string{
			"linkToken": linkToken,
		})
	})

	router.POST("/generateAccessToken", func(c *gin.Context) {
		var req map[string]string
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		publicToken := req["publicToken"]
		plaidRepository.GetAccessToken(publicToken)
		c.JSON(http.StatusOK, gin.H{"public_token_exchange": "complete"})
	})

	return router.Run(fmt.Sprintf(":%d", port))
}

func returnErrorJson(err error, c *gin.Context) {
	fmt.Println(err.Error())
	c.AbortWithStatusJSON(500, gin.H{
		"error": err.Error(),
	})
}

func returnErrorJsonCode(err error, c *gin.Context, code int) {
	fmt.Println(err.Error())
	c.AbortWithStatusJSON(code, gin.H{
		"error": err.Error(),
	})
}

func blockBots(c *gin.Context) {
	clientIP := c.ClientIP()
	blockedIps := []string{"172.31.45.22"}
	for _, ip := range blockedIps {
		if ip == clientIP {
			c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
			c.Abort()
			return
		}
	}
	c.Next()
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
