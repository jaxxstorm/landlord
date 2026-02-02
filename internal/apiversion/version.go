package apiversion

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

const Current = "v1"

var Supported = []string{Current}

var versionSegmentPattern = regexp.MustCompile(`^v[0-9]+$`)

func IsSupported(version string) bool {
	for _, supported := range Supported {
		if supported == version {
			return true
		}
	}
	return false
}

func SupportedVersions() []string {
	return append([]string(nil), Supported...)
}

func NormalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return baseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" {
		return appendVersionIfMissing(baseURL)
	}

	lastSegment := path.Base(parsed.Path)
	if lastSegment == "." || lastSegment == "/" {
		lastSegment = ""
	}

	if lastSegment == "" {
		parsed.Path = "/" + Current
		return parsed.String()
	}

	if versionSegmentPattern.MatchString(lastSegment) {
		return parsed.String()
	}

	if lastSegment == "api" {
		parsed.Path = path.Join(path.Dir(parsed.Path), Current)
		return parsed.String()
	}

	parsed.Path = path.Join(parsed.Path, Current)
	return parsed.String()
}

func appendVersionIfMissing(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		return baseURL
	}

	parts := strings.Split(trimmed, "/")
	lastSegment := parts[len(parts)-1]
	if versionSegmentPattern.MatchString(lastSegment) {
		return trimmed
	}

	if lastSegment == "api" {
		parts[len(parts)-1] = Current
		return strings.Join(parts, "/")
	}

	return trimmed + "/" + Current
}
