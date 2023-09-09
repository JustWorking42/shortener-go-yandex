package urlgenerator

import (
	"bytes"
	"math/rand"
)

const (
	alphabet        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortLinkLength = 5
)

func CreateShortLink() string {
	var bufer bytes.Buffer
	for i := 0; i < shortLinkLength; i++ {
		bufer.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return bufer.String()
}
