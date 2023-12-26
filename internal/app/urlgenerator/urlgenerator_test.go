package urlgenerator

import (
	"testing"
)

func BenchmarkCreateShortLink(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateShortLink()
	}
}
