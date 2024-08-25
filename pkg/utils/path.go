package utils

import "net/url"

func Path(str string) string {
	parsed, err := url.Parse(str)
	if err != nil {
		return ""
	}

	return parsed.Path
}

func Host(str string) string {
	parsed, err := url.Parse(str)
	if err != nil {
		return ""
	}

	return parsed.Host
}

func Schema(str string) string {
	parsed, err := url.Parse(str)
	if err != nil {
		return ""
	}

	return parsed.Scheme
}
