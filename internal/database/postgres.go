package database

import (
	"fmt"
	"log"
	"os"

	"UTM_Tracker/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Client *gorm.DB
}

func NewPostgresDatabase() *Database {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.URLMapping{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return &Database{
		Client: db,
	}
}

func (d *Database) CreateURLMapping(mapping *models.URLMapping) error {
	return d.Client.Create(mapping).Error
}

func (d *Database) GetURLBySlug(slug string) (*models.URLMapping, error) {
	var urlMapping models.URLMapping
	result := d.Client.Where("slug = ?", slug).First(&urlMapping)

	if result.Error != nil {
		return nil, result.Error
	}

	return &urlMapping, nil
}
