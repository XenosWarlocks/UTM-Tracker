package tracking

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdvancedTracker struct {
	db *gorm.DB
}

type TrackingMetadata struct {
	gorm.Model
	Slug           string `gorm:"index"`
	TrackingID     string `gorm:"uniqueIndex"`
	IPNetwork      string
	CountryCode    string
	BrowserFamily  string
	OSFamily       string
	DeviceType     string
	FirstClickTime time.Time
	ClickCount     int
}

func NewAdvancedTracker(db *gorm.DB) *AdvancedTracker {
	// Auto-migrate advanced tracking model
	err := db.AutoMigrate(&TrackingMetadata{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate advanced tracking model: %v", err))
	}

	return &AdvancedTracker{db: db}
}

// GenerateTrackingID creates a cryptographically secure unique tracking identifier
func (at *AdvancedTracker) GenerateTrackingID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(b)
}

// ExtractNetworkInfo parses IP address to extract network information
func (at *AdvancedTracker) ExtractNetworkInfo(ipStr string) (string, string) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", ""
	}

	// Extract network for anonymization
	var network string
	var countryCode string

	// IPv4 handling
	if ipv4 := ip.To4(); ipv4 != nil {
		// Mask to /24 for IPv4
		network = ip.Mask(net.CIDRMask(24, 32)).String()
	} else {
		// Mask to /48 for IPv6
		network = ip.Mask(net.CIDRMask(48, 128)).String()
	}

	// TODO: Implement geolocation lookup for country code
	// This would typically involve an IP geolocation database
	// countryCode = lookupCountryCode(ip)

	return network, countryCode
}

// ParseUserAgent extracts browser, OS, and device information
func (at *AdvancedTracker) ParseUserAgent(userAgent string) (string, string, string) {
	userAgent = strings.ToLower(userAgent)

	// Basic user agent parsing (simplified)
	var browserFamily, osFamily, deviceType string

	// Browser detection
	switch {
	case strings.Contains(userAgent, "chrome"):
		browserFamily = "Chrome"
	case strings.Contains(userAgent, "firefox"):
		browserFamily = "Firefox"
	case strings.Contains(userAgent, "safari"):
		browserFamily = "Safari"
	case strings.Contains(userAgent, "edge"):
		browserFamily = "Edge"
	default:
		browserFamily = "Unknown"
	}

	// OS detection
	switch {
	case strings.Contains(userAgent, "windows"):
		osFamily = "Windows"
	case strings.Contains(userAgent, "mac"):
		osFamily = "macOS"
	case strings.Contains(userAgent, "linux"):
		osFamily = "Linux"
	case strings.Contains(userAgent, "android"):
		osFamily = "Android"
	case strings.Contains(userAgent, "ios"):
		osFamily = "iOS"
	default:
		osFamily = "Unknown"
	}

	// Device type detection
	switch {
	case strings.Contains(userAgent, "mobile"):
		deviceType = "Mobile"
	case strings.Contains(userAgent, "tablet"):
		deviceType = "Tablet"
	default:
		deviceType = "Desktop"
	}

	return browserFamily, osFamily, deviceType
}

// RecordAdvancedTracking logs detailed tracking information
func (at *AdvancedTracker) RecordAdvancedTracking(slug, ipAddress, userAgent string) error {
	// Generate unique tracking ID
	trackingID := at.GenerateTrackingID()

	// Extract network information
	networkInfo, countryCode := at.ExtractNetworkInfo(ipAddress)

	// Parse user agent
	browserFamily, osFamily, deviceType := at.ParseUserAgent(userAgent)

	// Create tracking metadata
	metadata := TrackingMetadata{
		Slug:           slug,
		TrackingID:     trackingID,
		IPNetwork:      networkInfo,
		CountryCode:    countryCode,
		BrowserFamily:  browserFamily,
		OSFamily:       osFamily,
		DeviceType:     deviceType,
		FirstClickTime: time.Now().UTC(),
		ClickCount:     1,
	}

	// Save to database
	return at.db.Create(&metadata).Error
}

// UpdateClickTracking increments click count for existing tracking
func (at *AdvancedTracker) UpdateClickTracking(trackingID string) error {
	return at.db.Model(&TrackingMetadata{}).
		Where("tracking_id = ?", trackingID).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}
