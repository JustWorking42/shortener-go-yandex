// Package urlgenerator provides functionality for generating short links.
package urlgenerator

import (
	"bytes"
	"math/rand"
)

const (
	// alphabet is the set of characters used to generate short links.
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// shortLinkLength is the length of generated short links.
	shortLinkLength = 5
)

// CreateShortLink generates a random short link.
func CreateShortLink() string {
	var buffer bytes.Buffer
	for i := 0; i < shortLinkLength; i++ {
		buffer.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return buffer.String()
}
