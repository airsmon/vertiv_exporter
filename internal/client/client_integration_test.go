package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestFetchDeviceDataLogsInAndRetriesAfterUnauthorized(t *testing.T) {
	var (
		loginRequests atomic.Int32
		dataRequests  atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/login.cgi":
			loginRequests.Add(1)
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse login form: %v", err)
				http.Error(w, "invalid form", http.StatusBadRequest)
				return
			}
			if got, want := r.Form.Get("user_name"), encodeCredential("admin"); got != want {
				t.Errorf("encoded username = %q, want %q", got, want)
			}
			if got, want := r.Form.Get("user_password"), encodeCredential("secret"); got != want {
				t.Errorf("encoded password = %q, want %q", got, want)
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "authenticated", Path: "/"})
			w.WriteHeader(http.StatusOK)

		case "/cgi-bin/p05_equip_sample.cgi":
			dataRequests.Add(1)
			if got, want := r.URL.Query().Get("_equipId"), "23"; got != want {
				t.Errorf("_equipId = %q, want %q", got, want)
			}
			if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "authenticated" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte("3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.6,℃"))

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, false, "admin", "secret", false)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	samples, err := client.FetchDeviceData(context.Background(), 23)
	if err != nil {
		t.Fatalf("FetchDeviceData returned error: %v", err)
	}
	if got, want := samples[2].Value, 28.6; got != want {
		t.Fatalf("field 2 value = %v, want %v", got, want)
	}
	if got, want := loginRequests.Load(), int32(1); got != want {
		t.Fatalf("login requests = %d, want %d", got, want)
	}
	if got, want := dataRequests.Load(), int32(2); got != want {
		t.Fatalf("data requests = %d, want %d", got, want)
	}
}
