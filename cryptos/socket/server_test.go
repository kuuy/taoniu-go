package socket

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthorizationHeaderParsing(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		wsProtocol    string
		queryString   string
		expectedToken string
	}{
		{
			name:          "TAONIU scheme prefix",
			authHeader:    "TAONIU my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "Bearer scheme prefix",
			authHeader:    "Bearer my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "bearer lowercase scheme prefix",
			authHeader:    "bearer my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "Raw token in Authorization header",
			authHeader:    "my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "Sec-WebSocket-Protocol header fallback",
			wsProtocol:    "TAONIU my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "URL query access_token fallback",
			queryString:   "?access_token=my-jwt-token-123",
			expectedToken: "my-jwt-token-123",
		},
		{
			name:          "No token provided",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/"
			if tt.queryString != "" {
				url += tt.queryString
			}
			req := httptest.NewRequest("GET", url, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.wsProtocol != "" {
				req.Header.Set("Sec-WebSocket-Protocol", tt.wsProtocol)
			}

			token := ""
			bearer := req.Header.Get("Authorization")
			if bearer == "" {
				bearer = req.Header.Get("Sec-WebSocket-Protocol")
			}
			if bearer != "" {
				fields := strings.Fields(bearer)
				if len(fields) >= 2 && (strings.EqualFold(fields[0], "TAONIU") || strings.EqualFold(fields[0], "BEARER")) {
					token = fields[1]
				} else if len(fields) == 1 && !strings.EqualFold(fields[0], "TAONIU") && !strings.EqualFold(fields[0], "BEARER") {
					token = fields[0]
				} else if len(bearer) > 7 && (strings.EqualFold(bearer[:6], "TAONIU") || strings.EqualFold(bearer[:6], "BEARER")) {
					token = strings.TrimSpace(bearer[7:])
				} else {
					token = strings.TrimSpace(bearer)
				}
			}
			if token == "" {
				token = req.URL.Query().Get("access_token")
			}

			if token != tt.expectedToken {
				t.Errorf("expected token %q, got %q", tt.expectedToken, token)
			}
		})
	}
}
