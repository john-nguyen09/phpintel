package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
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

func GetURIID(uri string) string {
	u, err := url.Parse(uri)
	h := md5.New()
	io.WriteString(h, uri)
	hash := hex.EncodeToString(h.Sum(nil))
	// Fallback to MD5 only if uri cannot be parsed
	if err != nil {
		return hash
	}
	parts := strings.Split(u.Path, "/")
	parts, lastPart := parts[:len(parts)-1], parts[len(parts)-1]
	result := ""
	for _, part := range parts[0 : len(parts)-1] {
		r := []rune(part)
		if len(r) == 0 {
			continue
		}
		result += string(r[0])
	}
	result += "-" + lastPart + "-" + hash[:8]
	return result
}
