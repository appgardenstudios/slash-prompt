package main

import (
	"path/filepath"
)

const (
	// URIScheme is the scheme used for resource URIs
	URIScheme = "file://"
)

// BuildResourceURI constructs a resource URI from repo and path components
func BuildResourceURI(repo, path string) string {
	key := filepath.Join(repo, path)
	return URIScheme + key
}
