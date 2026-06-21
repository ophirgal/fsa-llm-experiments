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
		// Add routes here
	}

	return r
}
