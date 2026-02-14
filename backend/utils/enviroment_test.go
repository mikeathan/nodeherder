package utils_test

import (
	"node-herder/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJwtSecretKey(t *testing.T) {
	testCases := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{name: "missing secret", secret: "", wantErr: true},
		{name: "valid secret", secret: "test-secret", wantErr: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET_KEY", testCase.secret)

			secret, err := utils.GetJwtSecretKey()
			if testCase.wantErr {
				assert.Error(t, err)
				assert.Nil(t, secret)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, []byte(testCase.secret), secret)
		})
	}
}

func TestValidateServiceCredentials(t *testing.T) {
	testCases := []struct {
		name         string
		clients      string
		secrets      map[string]string
		clientID     string
		clientSecret string
		expected     bool
	}{
		{
			name:         "missing clients",
			clients:      "",
			clientID:     "service-a",
			clientSecret: "secret-a",
			expected:     false,
		},
		{
			name:         "client listed but secret missing",
			clients:      "service-a",
			clientID:     "service-a",
			clientSecret: "secret-a",
			expected:     false,
		},
		{
			name:         "client listed but secret mismatch",
			clients:      "service-a",
			secrets:      map[string]string{"service-a": "secret-a"},
			clientID:     "service-a",
			clientSecret: "wrong",
			expected:     false,
		},
		{
			name:         "matches configured client",
			clients:      "service-a",
			secrets:      map[string]string{"service_a": "secret-a"},
			clientID:     "service-a",
			clientSecret: "secret-a",
			expected:     true,
		},
		{
			name:         "trims client ids and matches second entry",
			clients:      " service-a , service-b ",
			secrets:      map[string]string{"service_b": "secret-b"},
			clientID:     "service-b",
			clientSecret: "secret-b",
			expected:     true,
		},
		{
			name:         "handles client id with hyphens by mapping to underscores",
			clients:      "llm-proxy",
			secrets:      map[string]string{"llm_proxy": "secret-proxy"},
			clientID:     "llm-proxy",
			clientSecret: "secret-proxy",
			expected:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("SERVICE_CLIENTS", testCase.clients)
			for id, secret := range testCase.secrets {
				t.Setenv("SERVICE_SECRET_"+id, secret)
			}

			ok := utils.ValidateServiceCredentials(testCase.clientID, testCase.clientSecret)
			assert.Equal(t, testCase.expected, ok)
		})
	}
}
