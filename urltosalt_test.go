package main

import (
	"bytes"
	"testing"
)

func TestUrlToSalt(t *testing.T) {
	urlTests := []struct {
		url  string
		salt string
	}{
		{"https://google.com", "google.com"},       // Simple.
		{"http://google.com", "http://google.com"}, // No-secure urls get no-secure prefix.
		{"https://google.com:8043", "google.com"},  // Port numbers are ignored.
		{"http://google.com:8080", "http://google.com"},

		{"https://subdomain.google.com", "google.com"}, // Don't include subdomain.
		{"http://subdomain.google.com", "http://google.com"},
		{"https://subdomain.github.io", "subdomain.github.io"}, // Except if it is a public suffix.
		{"http://subdomain.github.io", "http://subdomain.github.io"},
		{"https://our.warwick.ac.uk", "warwick.ac.uk"},
		{"http://www.accommodation.manchester.ac.uk", "http://manchester.ac.uk"}, // Shame on mcr https not covering everything lol

		{"ftp://example.com", "ftp://example.com"}, // Protocol passthrough
		{"ftp://subdomain.example.com", "ftp://example.com"},
		{"ftp://subdomain.github.io", "ftp://subdomain.github.io"},
	}
	errorTests := []string{
		"some-name.com",
		"invalid://",
		"https://com",
		"some-name.com/some/path",
		"/some-name.com/some/path",
		"name",
	}
	test := func(url, wantSalt string) {
		t.Run(url, func(t *testing.T) {
			got, err := UrlToSalt(url)
			if err != nil {
				t.Errorf("UrlToSalt() errored: %v", err)
				return
			}
			if bytes.Compare([]byte(wantSalt), got) != 0 {
				t.Errorf("UrlToSalt() = %v, want %v", string(got), wantSalt)
			}
		})
	}
	for _, tt := range urlTests {
		test(tt.url, tt.salt)
		test(tt.url+"/some/path?query", tt.salt)
	}
	for _, str := range errorTests {
		t.Run(str+" (should error)", func(t *testing.T) {
			_, err := UrlToSalt(str)
			if err == nil {
				t.Errorf("Expected some error, but got none.")
			}
		})
	}
}
