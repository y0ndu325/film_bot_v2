package service

import (
	"errors"
	"filmBot/internal/database"
	"filmBot/internal/models"
	"math/rand"
	"time"
)

var (
	ErrNoMovies     = errors.New("no movies in database")
	ErrInvalidIndex = errors.New("invalid movie index")
)

type MovieService struct {
	db *database.Database
}

func New(db *database.Database) *MovieService {
	return &MovieService{db: db}
}

func (s *MovieService) AddMovie(title string) error {
	movie := &models.Movie{Title: title}
	return s.db.AddMovie(movie)
}

func (s *MovieService) GetRandomMovie() (*models.Movie, error) {
	movies, err := s.db.GetMovies()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(movies))
	return &movies[randomIndex], nil
}

func (s *MovieService) GetMovies() ([]models.Movie, error) {
	return s.db.GetMovies()
}

func (s *MovieService) DeleteMovie(index int) (*models.Movie, error) {
	movies, err := s.db.GetMovies()
	if err != nil {
		return nil, err
	}
	if index < 1 || index > len(movies) {
		return nil, ErrInvalidIndex
	}

	movie := movies[index-1]
	if err := s.db.DeleteMovie(&movie); err != nil {
		return nil, err
	}
	return &movie, nil
}
