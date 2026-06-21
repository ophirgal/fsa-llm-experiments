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
	}

	return r
}
