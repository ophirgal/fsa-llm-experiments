package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fsa-llm-experiments/backend/dal"
	"fsa-llm-experiments/backend/internal/service"
)

func ListExperiments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		experiments, err := dal.ListExperiments(c.Request.Context(), db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, experiments)
	}
}

func GetExperiment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		experiment, err := dal.GetExperiment(c.Request.Context(), db, id)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "experiment not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, experiment)
	}
}

func CreateExperiment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Name        string   `json:"name"`
			DatasetID   int64    `json:"dataset_id"`
			Prompts     []string `json:"prompts"`
			JudgePrompt string   `json:"judge_prompt"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if body.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if body.DatasetID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dataset_id is required"})
			return
		}
		if len(body.Prompts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one prompt is required"})
			return
		}
		if body.JudgePrompt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "judge_prompt is required"})
			return
		}

		created, err := dal.CreateExperiment(c.Request.Context(), db, body.Name, body.DatasetID, body.Prompts, body.JudgePrompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if c.Query("isStarted") == "true" {
			started, err := dal.StartExperiment(c.Request.Context(), db, created.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created.Experiment = started
			go service.RunExperiment(context.Background(), db, created.ID, nil)
		}

		c.JSON(http.StatusCreated, created)
	}
}

func UpdateExperiment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if body.Status != "in progress" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only status 'in progress' is accepted"})
			return
		}

		experiment, err := dal.StartExperiment(c.Request.Context(), db, id)
		if errors.Is(err, dal.ErrAlreadyStarted) {
			c.JSON(http.StatusConflict, gin.H{"error": "experiment already started"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		go service.RunExperiment(context.Background(), db, id, nil)

		c.JSON(http.StatusOK, experiment)
	}
}
