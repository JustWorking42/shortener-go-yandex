package cookie

import (
	"testing"
)

func BenchmarkGetUserID(b *testing.B) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyX2lkIn0.qgwhzie6gvs8BiiUfGSuODdJSr4cOmR7pggYrG3bT78"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getUserID(tokenString)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	userID := "user_id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateToken(userID)
	}
}
