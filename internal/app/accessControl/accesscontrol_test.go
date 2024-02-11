package accesscontrol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestCidrAccessMidlware(t *testing.T) {

	tests := []struct {
		name           string
		ip             string
		cidr           string
		expectedStatus int
	}{
		{
			name:           "AllowedIp",
			ip:             "192.168.1.1",
			cidr:           "192.168.1.0/24",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DeniedIp",
			ip:             "192.168.1.1",
			cidr:           "193.168.1.0/24",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "EmptyCidr",
			ip:             "192.168.1.1",
			cidr:           "",
			expectedStatus: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			app, err := app.CreateApp(context.Background(), configs.Config{
				TrustedSubnet: test.cidr,
			})
			assert.NoError(t, err)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server := httptest.NewServer(CidrAccessMiddleware(app, handler))
			defer server.Close()

			client := resty.New()

			resp, err := client.R().SetHeader("X-Real-IP", test.ip).Get(server.URL)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode())
		})
	}
}
