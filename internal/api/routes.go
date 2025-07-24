package api

import (
	apiHandler "dns-server/internal/api/apiHandlers"

	"github.com/gin-gonic/gin"
)

func HandleFuncs(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/records", apiHandler.GetRecords)
		api.POST("/records", apiHandler.CreateRecord)
		api.DELETE("/records/:domain", apiHandler.DeleteRecord)
		api.GET("/health", apiHandler.HealthCheck)
	}

}
