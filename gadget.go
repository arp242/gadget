package gadget

import (
	"fmt"
	"strings"
)

// Assume lowest Safari version.
// https://en.wikipedia.org/wiki/Safari_version_history#Safari_11_2

var (
	// Map Windows NT versions to actual product versions.
	windowsVersions = map[string]string{"5.0": "2000", "5.1": "XP", "5.2": "XP",
		"6.0": "Vista", "6.1": "7", "6.2": "8", "6.3": "8.1", "10.0": "10"}

	// https://en.wikipedia.org/wiki/Safari_version_history#Safari_10
	safariVersions = map[string]string{
		"601.1.46":  "9.0",
		"601.1.56":  "9.0",
		"601.2.7":   "9.0",
		"601.3.9":   "9.0",
		"601.4.4":   "9.0",
		"601.5.17":  "9.1",
		"601.6.17":  "9.1",
		"601.7.1":   "9.1",
		"601.7.8":   "9.1",
		"602.1.50 ": "10.0",
		"602.1.50":  "10.0",
		"602.2.14":  "10.0",
		"602.3.12":  "10.0",
		"602.4.6 ":  "10.0",
		"602.4.8":   "10.0",
		"603.1.30":  "10.1",
		"603.2.4 ":  "10.0",
		"603.2.4":   "10.1",
		"603.3.8":   "10.1",
		"604.1.38":  "11.0",
		"604.2.4":   "11.0",
		"604.3.5":   "11.0",
		"604.4.7":   "11.0",
		"604.5.6":   "11.0",
		"605.1.15":  "11.0",
		"605.1.33":  "11.1",
		"606.1.36":  "12.0",
		"607.1.40":  "12.1",
		"608.2.11":  "13.0",
	}

	// Meaningless product strings to ignore.
	ignoreProduct = []string{"Mozilla/", "Gecko/", "AppleWebKit/",
		"(KHTML,", "like", "Gecko)"}

	// Known browsers that are not based on Chrome, Safari, or Firefox.
	knownBrowsers = []string{
		"BingPreview/",
		"PhantomJS/",
	}
)

type UserAgent struct {
	BrowserName    string
	BrowserVersion string
	OSName         string
	OSVersion      string
}

func (u UserAgent) String() string {
	if u.OS() == "" {
		return u.Browser()
	}
	if u.Browser() == "" {
		return u.OS()
	}
	return fmt.Sprintf("%s on %s", u.Browser(), u.OS())
}

func (u UserAgent) Browser() string {
	if u.BrowserVersion == "" {
		return u.BrowserName
	}
	return fmt.Sprintf("%s %s", u.BrowserName, u.BrowserVersion)
}

func (u UserAgent) OS() string {
	if u.OSVersion == "" {
		return u.OSName
	}
	return fmt.Sprintf("%s %s", u.OSName, u.OSVersion)
}

func Parse(uaHeader string) UserAgent {
	p := parse(uaHeader)
	ua := UserAgent{}

	// Get OS info.
	isIE := false
	{
	oloop:
		for _, s := range p.system {
			switch {
			// Special IE stuff...
			case strings.HasPrefix(s, "Trident/") || strings.HasPrefix(s, "MSIE "):
				isIE = true

			case strings.HasPrefix(s, "Linux") || s == "X11":
				ua.OSName = "Linux"
				// Don't break as this might be Android.

			case strings.HasPrefix(s, "Android"):
				ua.OSName = "Android"
				if len(s) > 7 {
					ua.OSVersion = maxVersion(s[8:], 2, true)
				}
				break oloop

			case strings.HasPrefix(s, "Intel Mac OS X"):
				ua.OSName = "macOS"
				if len(s) > 14 {
					ua.OSVersion = maxVersion(strings.ReplaceAll(s[15:], "_", "."), 2, false)
				}
				break oloop

			case strings.HasPrefix(s, "CPU iPhone OS") || strings.HasPrefix(s, "CPU OS"):
				ua.OSName = "iOS"
				i := strings.Index(s, "OS")
				if i > -1 && len(s) > i+3 {
					// CPU iPhone OS 7_0 like Mac OS X
					v := s[i+3:]
					v = v[:strings.IndexRune(v, ' ')]
					ua.OSVersion = maxVersion(strings.ReplaceAll(v, "_", "."), 2, false)
				}
				break oloop

			case strings.HasPrefix(s, "Windows"):
				ua.OSName = "Windows"
				if i := strings.LastIndexByte(s, ' '); i > -1 {
					ua.OSVersion = windowsVersions[s[i+1:]]
				}
				// Don't break to detect Trident/ string for IE 11
				//break oloop
			}
		}
	}

	if isIE {
		ua.BrowserName = "Internet Explorer"
		if i := strings.Index(uaHeader, "MSIE "); i > -1 {
			ua.BrowserVersion = string(uaHeader[i+5])
		} else {
			ua.BrowserVersion = "11"
		}
		return ua
	}

	// Get browser info.
	{
		for _, s := range p.products {
			for _, k := range knownBrowsers {
				if strings.HasPrefix(s, k) {
					slash := strings.IndexRune(s, '/')
					if slash > -1 {
						ua.BrowserName = s[:slash]
						ua.BrowserVersion = s[slash+1:]
					}

					return ua
				}
			}
		}

	bloop:
		for _, s := range p.products {
			for _, ig := range ignoreProduct {
				if s == ig {
					continue bloop
				}
			}

			switch {
			case strings.HasPrefix(s, "Chrome/"):
				// EdgeHTML identifies as Chrome, even though it's not.
				for _, s2 := range p.products {
					if strings.HasPrefix(s2, "Edge/") {
						v := maxVersion(s2[5:], 1, false)
						if v[0] == '1' {
							ua.BrowserName = "Edge"
							ua.BrowserVersion = v
						}
						break bloop
					}
				}

				ua.BrowserName = "Chrome"
				ua.BrowserVersion = maxVersion(s[7:], 1, false)
				break bloop

			case strings.HasPrefix(s, "HeadlessChrome/"):
				ua.BrowserName = "Chrome"
				ua.BrowserVersion = maxVersion(s[15:], 1, false)
				break bloop

			case strings.HasPrefix(s, "Firefox/"):
				ua.BrowserName = "Firefox"
				ua.BrowserVersion = maxVersion(s[8:], 1, false)
				break bloop

			// We need to do all sort of tricks for Safari :-/
			case (ua.OSName == "macOS" || ua.OSName == "iOS") && (strings.HasPrefix(s, "Safari/") || s == "Mobile/15E148"):
				var (
					isSafari bool
					version  string
					webkit   string
				)
				for _, s2 := range p.products {
					if strings.HasPrefix(s2, "Version/") && !strings.Contains(s2, "15E") {
						version = s2[8:]
						isSafari = true
					}
					if strings.HasPrefix(s2, "AppleWebKit/") {
						webkit = s2
					}

					if s2 == "Mobile/15E148" {
						isSafari = true
					}
					if strings.HasPrefix(s2, "FxiOS/") || strings.HasPrefix(s2, "CriOS/") || strings.HasPrefix(s2, "Mobile/") {
						isSafari = true
					}
				}
				if isSafari {
					ua.BrowserName = "Safari"
					if version != "" {
						ua.BrowserVersion = maxVersion(version, 2, false)
					} else if webkit != "" {
						ua.BrowserVersion = safariVersions[webkit[12:]]
					}
					break bloop
				}
			}
		}

		// TODO: maybe look at system, too?
		if ua.BrowserName == "" {
			first := p.products[0]

			// Give up.
			// TODO: maybe try going forward?
			for _, ig := range ignoreProduct {
				if strings.HasPrefix(first, ig) {
					return ua
				}
			}

			s := strings.IndexRune(first, '/')
			if s > 0 && s < len(first)-1 && isNumber(first[s+1]) && isLetter(first[s-1]) {
				ua.BrowserName = first[:s]
				ua.BrowserVersion = first[s+1:]
			}
		}
	}

	return ua
}

func isNumber(s byte) bool { return s >= 0x30 && s <= 0x39 }
func isLetter(s byte) bool { return (s >= 0x41 && s <= 0x5a) || (s >= 0x61 && s <= 0x7a) }

// Set maximum version level:
//
// n=1: 75.0  -> 75
// n=2: 5.0.5 -> 5.0
func maxVersion(v string, n int, trimZero bool) string {
	if strings.Count(v, ".") >= n {
		s := strings.Split(v, ".")
		if trimZero && s[n-1] == "0" {
			n -= 1
		}
		return strings.Join(s[:n], ".")
	}

	if trimZero && strings.HasSuffix(v, ".0") {
		return v[:len(v)-2]
	}
	return v
}

type props struct {
	system   []string // System information between (..)
	products []string // All the Foo/ver products
}

func parse(ua string) props {
	p := props{}

	s := strings.IndexRune(ua, '(')
	e := strings.IndexRune(ua, ')')
	if e < s {
		e = -1
		s = -1
	}
	if s > -1 && e > -1 {
		p.system = strings.Split(ua[s+1:e], ";")
		for i := range p.system {
			p.system[i] = strings.TrimSpace(p.system[i])
		}
	}

	if e > -1 && s > -1 {
		p.products = strings.Split(
			strings.TrimSpace(ua[:s])+" "+strings.TrimSpace(ua[e+1:]), " ")
	} else {
		p.products = strings.Split(ua, " ")
	}
	for i := range p.products {
		p.products[i] = strings.TrimSpace(p.products[i])
	}

	return p
}
