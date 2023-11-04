package api

import (
	"bytes"
	"fmt"
	"hood/internal/resolver"
	"net/http"

	api "github.com/sahilsk11/ace-common/types"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartApi(port int, r resolver.Resolver) error {
	router := gin.Default()

	router.Use(blockBots)
	router.Use(cors.Default())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{"message": "welcome to hood"})
	})

	router.POST("/portfolioCorrelation", func(ctx *gin.Context) {
		var req api.CorrelationMatrixRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), ctx)
			return
		}

		resp, err := r.PortfolioCorrelation(req)
		if err != nil {
			returnErrorJson(err, ctx)
			return
		}

		ctx.JSON(200, resp)
	})

	router.POST("/correlatedAssetGroups", func(ctx *gin.Context) {
		var req api.CorrelatedAssetGroupsRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			returnErrorJson(fmt.Errorf("failed to read request body: %w", err), ctx)
			return
		}

		resp, err := r.CorrelatedAssetGroups(req)
		if err != nil {
			returnErrorJson(err, ctx)
			return
		}

		ctx.JSON(200, resp)
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
