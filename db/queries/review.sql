-- Review queries

-- name: CreateReview :one
INSERT INTO task_reviews (task_id, reviewer_id, status, submission_note)
VALUES ($1, $2, $3, $4)
RETURNING id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at;

-- name: GetReviewByID :one
SELECT id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at
FROM task_reviews
WHERE id = $1 AND task_id = $2;

-- name: ListReviews :many
SELECT id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at
FROM task_reviews
WHERE task_id = $1
ORDER BY created_at DESC;

-- name: ApproveReview :one
UPDATE task_reviews
SET status = 'approved', review_note = $3, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND task_id = $2 AND status = 'pending'
RETURNING id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at;

-- name: RequestChangesReview :one
UPDATE task_reviews
SET status = 'changes_requested', review_note = $3, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND task_id = $2 AND status = 'pending'
RETURNING id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at;
