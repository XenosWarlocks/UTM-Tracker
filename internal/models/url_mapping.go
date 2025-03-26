package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

// URLMapping represents the database model for storing short URL mappings
type URLMapping struct {
	gorm.Model
	Slug           string    `gorm:"uniqueIndex;not null"` // Unique slug for the short URL
	DestinationURL string    `gorm:"not null"`             // Original destination URL
	UTMSource      string    // UTM source parameter
	UTMMedium      string    // UTM medium parameter
	UTMCampaign    string    // UTM campaign parameter
	ExpiresAt      time.Time // Expiration date of the short URL
}

// Validate performs validation checks on the URLMapping
func (m *URLMapping) Validate() error {
	// Check if slug is empty
	if strings.TrimSpace(m.Slug) == "" {
		return errors.New("slug cannot be empty")
	}

	// Validate destination URL
	if _, err := url.ParseRequestURI(m.DestinationURL); err != nil {
		return errors.New("invalid destination URL")
	}

	// Validate UTM parameters if they exist
	if m.UTMSource != "" && !isValidUTMParam(m.UTMSource) {
		return errors.New("invalid UTM source")
	}

	if m.UTMMedium != "" && !isValidUTMParam(m.UTMMedium) {
		return errors.New("invalid UTM medium")
	}

	if m.UTMCampaign != "" && !isValidUTMParam(m.UTMCampaign) {
		return errors.New("invalid UTM campaign")
	}

	// Check expiration (optional)
	if !m.ExpiresAt.IsZero() && m.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration time must be in the future")
	}

	return nil
}

// isValidUTMParam checks if a UTM parameter is valid
func isValidUTMParam(param string) bool {
	// Simple validation: no special characters, max length
	return len(param) > 0 && len(param) <= 100 && strings.IndexFunc(
		param, func(r rune) bool {
			return !(r == '_' || r == '-' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
		}) == -1
}

// CreateURLMapping generates a new short URL mapping
func (m *URLMapping) CreateURLMapping(db *gorm.DB) error {
	// Validate the mapping
	if err := m.Validate(); err != nil {
		return err
	}

	// Check if slug already exists
	var existingMapping URLMapping
	result := db.Where("slug = ?", m.Slug).First(&existingMapping)
	if result.Error == nil {
		return errors.New("slug already exists")
	}

	// If error is not "record not found", return the error
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Set default expiration to 1 year from now if not set
	if m.ExpiresAt.IsZero() {
		m.ExpiresAt = time.Now().AddDate(1, 0, 0)
	}

	// Create the new mapping
	return db.Create(m).Error
}

// GetURLBySlug retrieves a URL mapping by its slug
func GetURLBySlug(db *gorm.DB, slug string) (*URLMapping, error) {
	// Trim and validate slug
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	var mapping URLMapping
	result := db.Where("slug = ? AND expires_at > ?", slug, time.Now()).First(&mapping)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("URL mapping not found or expired")
		}
		return nil, result.Error
	}

	return &mapping, nil
}

// BuildRedirectURL constructs the full redirect URL with UTM parameters
func (m *URLMapping) BuildRedirectURL() string {
	// Parse the base URL
	baseURL, err := url.Parse(m.DestinationURL)
	if err != nil {
		return m.DestinationURL
	}

	// Prepare query parameters
	query := baseURL.Query()

	// Add UTM parameters if they exist
	if m.UTMSource != "" {
		query.Set("utm_source", m.UTMSource)
	}
	if m.UTMMedium != "" {
		query.Set("utm_medium", m.UTMMedium)
	}
	if m.UTMCampaign != "" {
		query.Set("utm_campaign", m.UTMCampaign)
	}

	// Set the modified query
	baseURL.RawQuery = query.Encode()

	return baseURL.String()
}
