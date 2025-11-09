package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/api/dto"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/service"
)

type FeedHandler struct {
	feedService service.FeedService
}

func NewFeedHandler(feedService service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// @Summary Get feed data
// @Description Get feed data with multiple sections (trending, popular, latest, nearby)
// @Tags Feed
// @Accept json
// @Produce json
// @Param request body dto.FeedRequest true "Feed request with sections"
// @Success 200 {object} dto.FeedResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /feed [post]
func (h *FeedHandler) GetFeed(c *gin.Context) {
	var req dto.FeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the request payload").
			Mark(ierr.ErrValidation))
		return
	}

	response, err := h.feedService.GetFeed(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Increment view count for a place
// @Description Increment the view count for a specific place
// @Tags Feed
// @Accept json
// @Produce json
// @Param id path string true "Place ID"
// @Success 204
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /places/{id}/view [post]
func (h *FeedHandler) IncrementViewCount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("place ID is required").
			WithHint("Please provide a valid place ID").
			Mark(ierr.ErrValidation))
		return
	}

	err := h.feedService.IncrementViewCount(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
