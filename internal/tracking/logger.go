package tracking

import (
	"log"
	"net"
	"time"

	"gorm.io/gorm"
)

type ClickTracker struct {
	db *gorm.DB
}

type ClickLog struct {
	gorm.Model
	Slug      string `gorm:"index"`
	IPAddress string
	Timestamp time.Time
	Referrer  string
	UserAgent string
}

func NewClickTracker(db *gorm.DB) *ClickTracker {
	// Auto-migrate click log model
	err := db.AutoMigrate(&ClickLog{})
	if err != nil {
		log.Fatalf("Failed to migrate click log: %v", err)
	}

	return &ClickTracker{db: db}
}

func (ct *ClickTracker) LogClick(slug, ipAddress, referrer, userAgent string) error {
	// Anonymize IP address
	ip := net.ParseIP(ipAddress)
	var anonymizedIP string
	if ip != nil {
		// Mask last octet for IPv4 or last 80 bits for IPv6
		if ip.To4() != nil {
			anonymizedIP = ip.Mask(net.CIDRMask(24, 32)).String()
		} else {
			anonymizedIP = ip.Mask(net.CIDRMask(48, 128)).String()
		}
	}

	clickLog := ClickLog{
		Slug:      slug,
		IPAddress: anonymizedIP,
		Timestamp: time.Now().UTC(),
		Referrer:  referrer,
		UserAgent: userAgent,
	}

	return ct.db.Create(&clickLog).Error
}
