package api

import (
	"bytes"
	"context"
	"fmt"
	"hood/internal/resolver"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	api_types "github.com/sahilsk11/ace-common/types/hood"
)

func StartApi(port int, r resolver.Resolver) error {
	router := gin.Default()

	router.Use(blockBots)
	router.Use(cors.Default())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{"message": "welcome to hood"})
	})

	router.POST("/generatePlaidLinkToken", func(c *gin.Context) {
		ctx := context.Background()

		var req api_types.GeneratePlaidLinkTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		response, err := r.GeneratePlaidLinkToken(ctx, req)
		if err != nil {
			returnErrorJson(err, c)
			return
		}

		c.JSON(200, response)
	})

	router.POST("/addPlaidBankItem", func(c *gin.Context) {
		ctx := context.Background()

		var req api_types.AddPlaidBankItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		err := r.AddPlaidBankItem(ctx, req)
		if err != nil {
			returnErrorJson(err, c)
			return
		}

		c.JSON(http.StatusOK, gin.H{"public_token_exchange": "complete"})
	})

	router.POST("/getHoldings", func(c *gin.Context) {
		var req api_types.GetTradingAccountHoldingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		resp, err := r.GetTradingAccountHoldings(req)
		if err != nil {
			returnErrorJson(err, c)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	router.POST("/newManualTradingAccount", func(c *gin.Context) {
		var req api_types.NewManualTradingAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		resp, err := r.NewManualTradingAccount(req)
		if err != nil {
			returnErrorJson(err, c)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	router.POST("/updatePosition", func(c *gin.Context) {
		var req api_types.UpdatePositionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), c)
			return
		}

		resp, err := r.UpdatePosition(req)
		if err != nil {
			returnErrorJson(err, c)
			return
		}

		c.JSON(http.StatusOK, resp)
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
