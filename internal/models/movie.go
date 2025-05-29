package models

type Movie struct {
	ID    uint   `gorm:"primaryKey"`
	Title string `gorm:"uniqueIndex"`
}
