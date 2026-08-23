package httpapi

import (
	"bytes"
	"example.com/chemical-plant-permit-service/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermitAPI(t *testing.T) {
	s := httptest.NewServer(New(store.New()))
	defer s.Close()
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{{"GET", "/api/v1/permits", "", 200}, {"POST", "/api/v1/permits/cp-501/status", `{"status":"suspended"}`, 200}, {"POST", "/api/v1/permits/cp-501/status", `{"status":"active"}`, 400}, {"POST", "/api/v1/permits/nope/status", `{"status":"closed"}`, 404}} {
		req, _ := http.NewRequest(tc.method, s.URL+tc.path, bytes.NewBufferString(tc.body))
		res, e := http.DefaultClient.Do(req)
		if e != nil || res.StatusCode != tc.want {
			t.Fatalf("%s %s: err=%v status=%d", tc.method, tc.path, e, res.StatusCode)
		}
		res.Body.Close()
	}
}
