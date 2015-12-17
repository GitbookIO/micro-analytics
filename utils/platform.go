package utils

import (
	"regexp"
)

func Platform(ua string) string {
	// Default value
	platform := ""

	// Map of Regexp
	platforms := map[*regexp.Regexp]string{
		regexp.MustCompile(`(?i)windows nt`):    "Microsoft Windows",
		regexp.MustCompile(`(?i)windows phone`): "Microsoft Windows Phone",
		regexp.MustCompile(`(?i)macintosh`):     "Apple Mac",
		regexp.MustCompile(`(?i)linux`):         "Linux",
		regexp.MustCompile(`(?i)wii`):           "Wii",
		regexp.MustCompile(`(?i)playstation`):   "Playstation",
		regexp.MustCompile(`(?i)ipad`):          "iPad",
		regexp.MustCompile(`(?i)ipod`):          "iPod",
		regexp.MustCompile(`(?i)iphone`):        "iPhone",
		regexp.MustCompile(`(?i)android`):       "Android",
		regexp.MustCompile(`(?i)blackberry`):    "Blackberry",
		regexp.MustCompile(`(?i)samsung`):       "Samsung",
		regexp.MustCompile(`(?i)curl`):          "Curl",
	}

	for regex, p := range platforms {
		if regex.MatchString(ua) {
			platform = p
			break
		}
	}

	return platform
}
