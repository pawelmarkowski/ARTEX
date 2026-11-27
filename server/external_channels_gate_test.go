package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExternalChannelsDisabledEnv(t *testing.T) {
	cases := []struct {
		notify string
		ext    string
		want   bool
	}{
		{"", "", false},
		{"0", "", false},
		{"false", "off", false},
		{"1", "", true},
		{"true", "", true},
		{"ON", "", true},
		{"yes", "", true},
		{"", "1", true},
		{"0", "true", true},
	}
	for _, tc := range cases {
		t.Setenv("ARTEX_DISABLE_NOTIFY", tc.notify)
		t.Setenv("ARTEX_DISABLE_EXTERNAL_CHANNELS", tc.ext)
		if got := ExternalChannelsDisabled(); got != tc.want {
			t.Errorf("ARTEX_DISABLE_NOTIFY=%q ARTEX_DISABLE_EXTERNAL_CHANNELS=%q: got %v, want %v",
				tc.notify, tc.ext, got, tc.want)
		}
	}
}

func TestExternalChannelsDisabledHTTP403(t *testing.T) {
	t.Setenv("ARTEX_DISABLE_NOTIFY", "1")
	key := []byte("0123456789abcdef0123456789abcdef")
	s := &Server{jwtKey: key}
	tok, err := signJWT(key)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	reqs := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/notify/channels", `{"name":"x","kind":"dingtalk","config":{"webhook":"https://e.com/h"}}`},
		{http.MethodPatch, "/api/notify/channels/1", `{"enabled":true}`},
		{http.MethodPost, "/api/notify/channels/1/test", ""},
		{http.MethodPost, "/api/notify/deliveries/1/retry", ""},
		{http.MethodPut, "/api/chat-bridges/telegram", `{"enabled":true}`},
		{http.MethodPost, "/api/chat-bridges/telegram/test", ""},
		{http.MethodPut, "/api/settings", `{"notify_enabled":true}`},
	}
	for _, tc := range reqs {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("%s %s: status=%d body=%s, want 403", tc.method, tc.path, w.Code, w.Body.String())
			continue
		}
		var out map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || !strings.Contains(out["error"].(string), "disabled") {
			t.Errorf("%s %s: unexpected 403 body=%s", tc.method, tc.path, w.Body.String())
		}
	}
}
