package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestWriteError(t *testing.T) {
// 	jsonMsg := ErrorMessageJSON{
// 		ErrorCode: INCIDENT_NOT_FOUND,
// 		Message:   ErrIncidentNotFound.Error(),
// 		RequestID: "req-inc-id-1111",
// 	}
// 	rec := httptest.NewRecorder()
// 	writeError(rec, 400, jsonMsg)

// 	if rec.Code != 400 {
// 		t.Errorf("Code expected %d, got %d", 400, rec.Code)
// 	}
// 	bodyGot := bytes.TrimSpace(rec.Body.Bytes())
// 	jsonMsgInJson, _ := json.Marshal(map[string]any{"error": jsonMsg})

// 	if !bytes.Equal(bodyGot, jsonMsgInJson) {
// 		t.Errorf("Body expected %s, got %s", jsonMsgInJson, bodyGot)
// 	}
// }
