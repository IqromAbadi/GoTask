package review

import "time"

// Review represents a task review entity.
type Review struct {
	ID             string     `json:"id"`
	TaskID         string     `json:"task_id"`
	ReviewerID     string     `json:"reviewer_id"`
	Status         string     `json:"status"`
	SubmissionNote *string    `json:"submission_note"`
	ReviewNote     *string    `json:"review_note"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ReviewResponse is the public DTO.
type ReviewResponse struct {
	ID             string  `json:"id"`
	TaskID         string  `json:"task_id"`
	ReviewerID     string  `json:"reviewer_id"`
	Status         string  `json:"status"`
	SubmissionNote *string `json:"submission_note"`
	ReviewNote     *string `json:"review_note"`
	ReviewedAt     *string `json:"reviewed_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ToResponse converts a Review to a ReviewResponse.
func ToResponse(r *Review) ReviewResponse {
	resp := ReviewResponse{
		ID:             r.ID,
		TaskID:         r.TaskID,
		ReviewerID:     r.ReviewerID,
		Status:         r.Status,
		SubmissionNote: r.SubmissionNote,
		ReviewNote:     r.ReviewNote,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
	}
	if r.ReviewedAt != nil {
		s := r.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &s
	}
	return resp
}

// SubmitReviewRequest is the DTO for submitting a review.
type SubmitReviewRequest struct {
	SubmissionNote string `json:"submission_note"`
}

// ApproveReviewRequest is the DTO for approving a review.
type ApproveReviewRequest struct {
	ReviewNote string `json:"review_note"`
}

// RequestChangesRequest is the DTO for requesting changes.
type RequestChangesRequest struct {
	ReviewNote string `json:"review_note" validate:"required"`
}
