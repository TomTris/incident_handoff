package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestWriteJson(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		status    int
// 		requestID string
// 		v         any
// 	}{
// 		{"200", 200, "test-inc-id-77", "777"},
// 		{"200", 200, "test-inc-id-78", map[string]string{"777": "999"}},
// 		{"400", 400, "test-inc-id-79", ErrorMessageJSON{
// 			ErrorCode: INCIDENT_NOT_FOUND,
// 			Message:   ErrIncidentNotFound.Error(),
// 			RequestID: "test-inc-id-79",
// 		}},
// 	}

// 	for _, test := range tests {
// 		rec := httptest.NewRecorder()
// 		writeJSON(rec, test.status, test.requestID, test.v)

// 		if rec.Header().Get("Content-Type") != "application/json" {
// 			t.Errorf(`rec.Header().Get("Content-Type") = %s, want %s`, rec.Header().Get("Content-Type"), "application/json")
// 		}
// 		if rec.Code != test.status {
// 			t.Errorf("http code wrong, expected %d, want %d", test.status, rec.Code)
// 		}
// 		expected, _ := json.Marshal(test.v)
// 		got := bytes.TrimSpace(rec.Body.Bytes())
// 		if !bytes.Equal(expected, got) {
// 			t.Errorf("body = %s, want %s", got, expected)
// 		}
// 	}

// 	t.Run("Fail", func(t *testing.T) {
// 		rec := httptest.NewRecorder()
// 		q := make(chan string, 1)
// 		writeJSON(rec, 400, "test-inc-id-85", q)

// 		if rec.Header().Get("Content-Type") != "application/json" {
// 			t.Errorf(`rec.Header().Get("Content-Type") = %s, want %s`, rec.Header().Get("Content-Type"), "application/json")
// 		}
// 		if rec.Code != 500 {
// 			t.Errorf("http code wrong, expected %d, want %d", 500, rec.Code)
// 		}
// 	})
// }
