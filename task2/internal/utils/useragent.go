package utils

import (
	"strings"

	"github.com/pozedorum/WB_project_3/task2/internal/models"
)

func ExtractBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "chromium"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "edge"):
		return "Edge"
	case strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "ie") || strings.Contains(ua, "trident"):
		return "Internet Explorer"
	case strings.Contains(ua, "brave"):
		return "Brave"
	default:
		return "Other"
	}
}

func ExtractOS(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		return "macOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "ubuntu"):
		return "Ubuntu"
	case strings.Contains(ua, "fedora"):
		return "Fedora"
	default:
		return "Other"
	}
}

func ExtractDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || strings.Contains(ua, "android"):
		return "Mobile"
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		return "Tablet"
	case strings.Contains(ua, "tv") || strings.Contains(ua, "smarttv"):
		return "TV"
	case strings.Contains(ua, "bot") || strings.Contains(ua, "crawler"):
		return "Bot"
	default:
		return "Desktop"
	}
}

func ParseUserAgent(userAgent string) models.UserAgentInfo {
	return models.UserAgentInfo{
		Browser: ExtractBrowser(userAgent),
		OS:      ExtractOS(userAgent),
		Device:  ExtractDevice(userAgent),
	}
}
