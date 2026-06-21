package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func New(db *sql.DB) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", Health)
		v1.GET("/datasets", ListDatasets(db))
		v1.POST("/datasets", CreateDataset(db))
		v1.GET("/experiments", ListExperiments(db))
		v1.GET("/experiments/:id", GetExperiment(db))
		v1.POST("/experiments", CreateExperiment(db))
		v1.PUT("/experiments/:id", UpdateExperiment(db))
	}

	return r
}
