package controllers

import (
	"github.com/Biskwit/langflow-openai-proxy/services"
	"github.com/Biskwit/langflow-openai-proxy/types"
	"github.com/gin-gonic/gin"
)

func ChatCompletions(c *gin.Context) {
	var body types.ChatCompletionsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatus(406)
		return
	}
	services.ChatCompletions(c, body)
}
