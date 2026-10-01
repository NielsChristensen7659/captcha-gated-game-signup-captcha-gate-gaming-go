package infrai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyCaptchaRequestBoundary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/captcha/verify" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input["widget_record_id"] != "widget-123" || input["token"] != "browser-token" || input["action"] != "game_signup" {
			t.Fatalf("request body = %#v", input)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{"valid":true},"error":null,"metadata":{}}`))
	}))
	defer server.Close()

	client := newClient(server.URL, "test-key", server.Client())
	err := client.VerifyCaptcha(context.Background(), CaptchaInput{WidgetRecordID: "widget-123", Token: "browser-token", Action: "game_signup", ScoreThreshold: 0.7})
	if err != nil {
		t.Fatal(err)
	}
}
