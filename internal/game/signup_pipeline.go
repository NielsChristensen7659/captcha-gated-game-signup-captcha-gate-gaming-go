package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"example.com/captcha-game-signup/internal/infrai"
)

type Registration struct {
	RegistrationID string `json:"registration_id"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	PlayerName     string `json:"player_name"`
	WidgetRecordID string `json:"widget_record_id"`
	CaptchaToken   string `json:"captcha_token"`
	CaptchaVendor  string `json:"captcha_vendor,omitempty"`
	IP             string `json:"-"`
}

type PlayerAsset struct {
	AssetID  string `json:"asset_id"`
	PlayerID string `json:"player_id"`
	Kind     string `json:"kind"`
	State    string `json:"state"`
}

type LiveEvent struct {
	EventID  string    `json:"event_id"`
	PlayerID string    `json:"player_id"`
	Name     string    `json:"name"`
	At       time.Time `json:"at"`
}

type ModerationItem struct {
	QueueID  string `json:"queue_id"`
	AssetID  string `json:"asset_id"`
	PlayerID string `json:"player_id"`
	Reason   string `json:"reason"`
	State    string `json:"state"`
}

type SignupResult struct {
	PlayerID string         `json:"player_id"`
	Asset    PlayerAsset    `json:"asset"`
	Event    LiveEvent      `json:"event"`
	Review   ModerationItem `json:"moderation"`
}

type Gateway interface {
	VerifyCaptcha(context.Context, infrai.CaptchaInput) error
	CreateUser(context.Context, infrai.UserInput) (infrai.User, error)
}

type Pipeline struct {
	gateway Gateway
	now     func() time.Time
	mu      sync.RWMutex
	results map[string]SignupResult
}

func NewPipeline(gateway Gateway) *Pipeline {
	return &Pipeline{gateway: gateway, now: time.Now, results: make(map[string]SignupResult)}
}

func (p *Pipeline) Signup(ctx context.Context, registration Registration) (SignupResult, error) {
	if registration.RegistrationID == "" || registration.Email == "" || registration.WidgetRecordID == "" || registration.CaptchaToken == "" {
		return SignupResult{}, &InputError{Message: "registration_id, email, widget_record_id, and captcha_token are required"}
	}

	p.mu.RLock()
	previous, found := p.results[registration.RegistrationID]
	p.mu.RUnlock()
	if found {
		return previous, nil
	}

	if err := p.gateway.VerifyCaptcha(ctx, infrai.CaptchaInput{
		WidgetRecordID: registration.WidgetRecordID, Token: registration.CaptchaToken, Vendor: registration.CaptchaVendor,
		IP: registration.IP, Action: "game_signup", ScoreThreshold: 0.7,
	}); err != nil {
		return SignupResult{}, err
	}

	user, err := p.gateway.CreateUser(ctx, infrai.UserInput{
		Email: registration.Email, Password: registration.Password, Name: registration.PlayerName,
		Metadata: map[string]any{"source": "game_signup"}, Mode: "player",
		IdempotencyKey: registration.RegistrationID,
	})
	if err != nil {
		return SignupResult{}, err
	}

	result := SignupResult{
		PlayerID: user.ID,
		Asset:    PlayerAsset{AssetID: "starter-" + registration.RegistrationID, PlayerID: user.ID, Kind: "player_banner", State: "pending_review"},
		Event:    LiveEvent{EventID: "signup-" + registration.RegistrationID, PlayerID: user.ID, Name: "player.signed_up", At: p.now().UTC()},
		Review:   ModerationItem{QueueID: "review-" + registration.RegistrationID, AssetID: "starter-" + registration.RegistrationID, PlayerID: user.ID, Reason: "new_player_asset", State: "queued"},
	}
	p.mu.Lock()
	p.results[registration.RegistrationID] = result
	p.mu.Unlock()
	return result, nil
}

type InputError struct{ Message string }

func (e *InputError) Error() string { return fmt.Sprintf("invalid signup: %s", e.Message) }
