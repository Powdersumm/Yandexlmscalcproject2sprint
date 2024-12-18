package application_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		expression     string
		expected       string
		expectedStatus int
	}{
		{"2+2", "result: 4.000000", http.StatusOK},
		{"2/0", "err: division by zero", http.StatusBadRequest},
		{"abc", "err: invalid expression", http.StatusBadRequest},
		{"", "EOF", http.StatusBadRequest},
		{"invalid", "internal server error", http.StatusInternalServerError},
	}

	for _, test := range tests {
		reqBody := `{"expression":"` + test.expression + `"}`

		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		CalcHandler(w, req)

		res := w.Result()

		if res.StatusCode != test.expectedStatus {
			t.Errorf("for expression %q: expected status %v, got %v", test.expression, test.expectedStatus, res.StatusCode)
		}

		body := w.Body.String()
		if body != test.expected {
			t.Errorf("for expression %q: expected body %q, got %q", test.expression, test.expected, body)
		}
	}
}
