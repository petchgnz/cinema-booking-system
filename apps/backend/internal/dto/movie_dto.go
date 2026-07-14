package dto

type CreateMovieRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Duration    int    `json:"duration" binding:"required,min=1"`
	PosterURL   string `json:"poster_url"`
}
