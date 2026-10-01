package game

import (
	"context"
	"errors"
	"testing"

	"example.com/captcha-game-signup/internal/infrai"
)

type fakeGateway struct {
	verifyErr   error
	verifyCalls int
	createCalls int
}

func (f *fakeGateway) VerifyCaptcha(context.Context, infrai.CaptchaInput) error {
	f.verifyCalls++
	return f.verifyErr
}

func (f *fakeGateway) CreateUser(context.Context, infrai.UserInput) (infrai.User, error) {
	f.createCalls++
	return infrai.User{ID: "player-42"}, nil
}

func TestSignupDecision(t *testing.T) {
	tests := []struct {
		name        string
		verifyErr   error
		wantCreated int
		wantErr     bool
	}{
		{name: "verified human enters player pipeline", wantCreated: 1},
		{name: "captcha rejection stops before account creation", verifyErr: &infrai.APIError{Message: "captcha was not accepted", HTTPStatus: 422}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := &fakeGateway{verifyErr: tt.verifyErr}
			pipeline := NewPipeline(gateway)
			result, err := pipeline.Signup(context.Background(), Registration{
				RegistrationID: "reg-123", Email: "player@example.com", WidgetRecordID: "widget-123", CaptchaToken: "browser-token",
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Signup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gateway.verifyCalls != 1 || gateway.createCalls != tt.wantCreated {
				t.Fatalf("calls verify=%d create=%d", gateway.verifyCalls, gateway.createCalls)
			}
			if !tt.wantErr && (result.PlayerID != "player-42" || result.Review.State != "queued") {
				t.Fatalf("unexpected pipeline result: %+v", result)
			}
			if tt.wantErr && !errors.Is(err, tt.verifyErr) {
				t.Fatalf("Signup() error = %v, want %v", err, tt.verifyErr)
			}
		})
	}
}

func TestSignupReturnsSameResultForRegistrationID(t *testing.T) {
	gateway := &fakeGateway{}
	pipeline := NewPipeline(gateway)
	registration := Registration{RegistrationID: "reg-123", Email: "player@example.com", WidgetRecordID: "widget-123", CaptchaToken: "browser-token"}

	first, err := pipeline.Signup(context.Background(), registration)
	if err != nil {
		t.Fatal(err)
	}
	second, err := pipeline.Signup(context.Background(), registration)
	if err != nil {
		t.Fatal(err)
	}
	if first.PlayerID != second.PlayerID || gateway.createCalls != 1 {
		t.Fatalf("duplicate signup created %d users", gateway.createCalls)
	}
}
