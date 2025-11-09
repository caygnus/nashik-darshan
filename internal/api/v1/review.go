package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/api/dto"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// ReviewHandler handles review-related HTTP requests
type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// @Summary Create a new review
// @Description Create a new review for a place or other entity
// @Tags Reviews
// @Accept json
// @Produce json
// @Param request body dto.CreateReviewRequest true "Create review request"
// @Success 201 {object} dto.ReviewResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews [post]
// @Security Authorization
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the request payload").
			Mark(ierr.ErrValidation))
		return
	}

	review, err := h.reviewService.CreateReview(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, review)
}

// @Summary Get review by ID
// @Description Get a review by its ID
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} dto.ReviewResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews/{id} [get]
func (h *ReviewHandler) GetReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("review ID is required").
			WithHint("Please provide a valid review ID").
			Mark(ierr.ErrValidation))
		return
	}

	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, review)
}

// @Summary Update a review
// @Description Update an existing review
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param request body dto.UpdateReviewRequest true "Update review request"
// @Success 200 {object} dto.ReviewResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews/{id} [put]
// @Security Authorization
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("review ID is required").
			WithHint("Please provide a valid review ID").
			Mark(ierr.ErrValidation))
		return
	}

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the request payload").
			Mark(ierr.ErrValidation))
		return
	}

	review, err := h.reviewService.UpdateReview(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, review)
}

// @Summary Delete a review
// @Description Delete a review
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Success 204
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews/{id} [delete]
// @Security Authorization
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("review ID is required").
			WithHint("Please provide a valid review ID").
			Mark(ierr.ErrValidation))
		return
	}

	err := h.reviewService.DeleteReview(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary List reviews
// @Description Get a paginated list of reviews with filtering
// @Tags Reviews
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param entity_type query string false "Entity type (place, experience, etc.)"
// @Param entity_id query string false "Entity ID"
// @Param user_id query string false "User ID"
// @Param min_rating query number false "Minimum rating"
// @Param max_rating query number false "Maximum rating"
// @Success 200 {object} types.ListResponse[dto.ReviewResponse]
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews [get]
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	filter := types.NewReviewFilter()

	// Bind query parameters to filter
	if err := c.ShouldBindQuery(filter); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the query parameters").
			Mark(ierr.ErrValidation))
		return
	}

	reviews, err := h.reviewService.ListReviews(c.Request.Context(), filter)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, reviews)
}

// @Summary Get rating statistics
// @Description Get rating statistics for an entity (place, experience, etc.)
// @Tags Reviews
// @Accept json
// @Produce json
// @Param entityType path string true "Entity type (place, experience, etc.)"
// @Param entityId path string true "Entity ID"
// @Success 200 {object} dto.RatingStatsResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /reviews/stats/{entityType}/{entityId} [get]
func (h *ReviewHandler) GetRatingStats(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")

	if entityType == "" {
		c.Error(ierr.NewError("entity type is required").
			WithHint("Please provide a valid entity type").
			Mark(ierr.ErrValidation))
		return
	}

	if entityID == "" {
		c.Error(ierr.NewError("entity ID is required").
			WithHint("Please provide a valid entity ID").
			Mark(ierr.ErrValidation))
		return
	}

	req := &dto.GetRatingStatsRequest{
		EntityType: types.ReviewEntityType(entityType),
		EntityID:   entityID,
	}

	stats, err := h.reviewService.GetRatingStats(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
