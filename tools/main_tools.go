package tools

import "github.com/gin-gonic/gin"

func GetToolsPage(c *gin.Context) {
	c.HTML(200, "tools.tmpl", gin.H{
		"PageTitle": "Тулзы",
		"CurPage":   "tools",
	})
}
