package validator

import "net/url"

func IsValidURL(checkableURL string) bool {
	u, err := url.Parse(checkableURL)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" && u.Host == "" {
		return false
	}

	return true
}
