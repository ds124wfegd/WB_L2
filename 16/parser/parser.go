package parser

import (
	"net/url"
	"strings"
)

func ExtractLinks(html string, base *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	patterns := []string{"href=\"", "src=\"", "action=\""}

	for _, pattern := range patterns {
		start := 0
		for {
			idx := strings.Index(html[start:], pattern)
			if idx == -1 {
				break
			}

			attrStart := start + idx + len(pattern)
			end := strings.Index(html[attrStart:], "\"")
			if end == -1 {
				break
			}

			link := html[attrStart : attrStart+end]

			if isValidLink(link) {
				absoluteLink := ResolveURL(link, base)
				if absoluteLink != "" && !seen[absoluteLink] {
					seen[absoluteLink] = true
					links = append(links, absoluteLink)
				}
			}

			start = attrStart + end + 1
		}
	}

	return links
}

func isValidLink(link string) bool {
	if link == "" {
		return false
	}

	invalidPrefixes := []string{
		"#", "javascript:", "mailto:", "data:", "tel:", "ftp:",
	}

	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(link, prefix) {
			return false
		}
	}

	return true
}

func ResolveURL(link string, base *url.URL) string {
	if link == "" {
		return ""
	}

	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	absolute, err := base.Parse(link)
	if err != nil {
		return ""
	}

	return absolute.String()
}

func IsSameDomain(link, baseURL string) bool {
	linkURL, err1 := url.Parse(link)
	base, err2 := url.Parse(baseURL)

	if err1 != nil || err2 != nil {
		return false
	}

	return linkURL.Host == base.Host
}
