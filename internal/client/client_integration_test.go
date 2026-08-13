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

func TestFetchDeviceDataLogsInAndRetriesAfterLoginRedirectBody(t *testing.T) {
	var (
		loginRequests atomic.Int32
		dataRequests  atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/login.cgi":
			loginRequests.Add(1)
			if _, err := r.Cookie("session"); err == nil {
				t.Error("login request retained the stale session cookie")
			}
			if cookie, err := r.Cookie("language"); err != nil || cookie.Value != "English" {
				t.Error("login request did not retain the language cookie")
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "authenticated", Path: "/"})
			w.WriteHeader(http.StatusOK)

		case "/cgi-bin/p05_equip_sample.cgi":
			dataRequests.Add(1)
			if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "authenticated" {
				_, _ = w.Write([]byte(`<html xmlns="http://www.w3.org/1999/xhtml"><head><script language="JavaScript" type="text/javascript">window.open("/cgi-bin/index.cgi", "_top");</script></head></html>`))
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
	client.httpClient.Jar.SetCookies(client.baseURL, []*http.Cookie{
		{Name: "session", Value: "stale", Path: "/"},
	})

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

func TestFetchDeviceDataRetriesLoginRedirectBodyOnlyOnce(t *testing.T) {
	var (
		loginRequests atomic.Int32
		dataRequests  atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/login.cgi":
			loginRequests.Add(1)
			w.WriteHeader(http.StatusOK)
		case "/cgi-bin/p05_equip_sample.cgi":
			dataRequests.Add(1)
			_, _ = w.Write([]byte(`<HTML><SCRIPT>WINDOW.OPEN('/cgi-bin/index.cgi', '_top');</SCRIPT></HTML>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, false, "admin", "secret", false)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if _, err := client.FetchDeviceData(context.Background(), 23); err == nil {
		t.Fatal("FetchDeviceData returned nil error for a persistent login redirect")
	}
	if got, want := loginRequests.Load(), int32(1); got != want {
		t.Fatalf("login requests = %d, want %d", got, want)
	}
	if got, want := dataRequests.Load(), int32(2); got != want {
		t.Fatalf("data requests = %d, want %d", got, want)
	}
}
