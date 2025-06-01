package database

import (
	"filmBot/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	if !db.Migrator().HasTable(&models.Movie{}) {
		if err := db.AutoMigrate(&models.Movie{}); err != nil {
			return nil, err
		}
	}
	return &Database{db: db}, nil
}

func (d *Database) AddMovie(movie *models.Movie) error {
	return d.db.Create(movie).Error
}

func (d *Database) GetMovies() ([]models.Movie, error) {
	var movies []models.Movie
	return movies, d.db.Find(&movies).Error
}

func (d *Database) DeleteMovie(movie *models.Movie) error {
	return d.db.Delete(movie).Error
}
