package api

import (
	"bufio"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"fsa-llm-experiments/backend/dal"
)

func ListDatasets(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		datasets, err := dal.ListDatasets(c.Request.Context(), db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, datasets)
	}
}

func CreateDataset(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		defer file.Close()

		var queries []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				queries = append(queries, line)
			}
		}
		if err := scanner.Err(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
			return
		}
		if len(queries) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file contains no queries"})
			return
		}

		created, err := dal.CreateDataset(c.Request.Context(), db, name, queries)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}
