package application

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		expression string
		expected   string
	}{
		{"3 4 +", "result: 7.000000"},
		{"10 2 -", "result: 8.000000"},
		{"5 5 *", "result: 25.000000"},
		{"8 2 /", "result: 4.000000"},
		{"invalid expression", "err: invalid expression"},
	}

	for _, test := range tests {
		reqBody := `{"expression":"` + test.expression + `"}`
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		CalcHandler(w, req)

		res := w.Result()
		if res.StatusCode != http.StatusOK && test.expected != "unknown err" {
			t.Errorf("expected status OK, got %v", res.Status)
		}

		body := w.Body.String()
		if body != test.expected {
			t.Errorf("expected %v, got %v", test.expected, body)
		}
	}
}
