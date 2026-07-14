package handler

import (
	"net/http"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/service"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	movieService service.MovieService
}

func NewMovieHandler(movieService service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

func (h *MovieHandler) Create(c *gin.Context) {
	// get request
	var req dto.CreateMovieRequest

	// parse json and validate like @Body in nest. if not pass, send bad req back
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	movie, err := h.movieService.CreateMovie(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
	}

	c.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) GetAll(c *gin.Context) {
	movies, err := h.movieService.GetAllMovies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	movie, err := h.movieService.GetMovieByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "movie not found"})
		return
	}

	c.JSON(http.StatusOK, movie)
}
