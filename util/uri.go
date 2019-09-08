package util

import (
	"errors"
	"net/url"
	"strings"
)

func UriToPath(uri string) string {
	urlIns, err := url.ParseRequestURI(uri)

	if err != nil {
		HandleError(err)
	}

	if urlIns.Scheme != "file" {
		HandleError(errors.New("Cannot convert non-file URI"))
	}
	filePath, err := url.QueryUnescape(urlIns.Path)

	if err != nil {
		HandleError(err)
	}

	if strings.Contains(filePath, ":") {
		if strings.HasPrefix(filePath, "/") {
			filePath = strings.TrimPrefix(filePath, "/")
		}
		filePath = strings.Replace(filePath, "/", "\\", -1)
	}

	return filePath
}

func PathToUri(path string) string {
	urlIns := new(url.URL)
	path = strings.Replace(path, "\\", "/", -1)
	parts := strings.Split(path, "/")
	first, parts := parts[0], parts[1:]

	if !strings.HasSuffix(first, ":") {
		first = url.QueryEscape(first)
	}

	tempParts := parts[:0]
	for _, part := range parts {
		tempParts = append(tempParts, url.QueryEscape(part))
	}
	parts = append([]string{first}, tempParts...)

	urlIns.Scheme = "file"
	urlIns.Path = "/" + strings.Join(parts, "/")

	return urlIns.String()
}
