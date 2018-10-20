package main

import (
	"fmt"
	"golang.org/x/net/publicsuffix"
	neturl "net/url"
)

// Extract the registerable part of the domain from the url and turn it into []byte.
// The result can then be passed to ecbpass.
func UrlToSalt(url string) (salt []byte, err error) {
	parsedUrl, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse url: %v", err.Error())
	}
	domain := parsedUrl.Hostname()
	publicDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, fmt.Errorf("Unable to look up public suffix for domain %v: %v", domain, err.Error())
	}
	if parsedUrl.Scheme == "https" {
		return []byte(publicDomain), nil
	} else {
		return []byte(fmt.Sprintf("%v://%v", parsedUrl.Scheme, publicDomain)), nil
	}
}
