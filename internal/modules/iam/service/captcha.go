package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type captchaState struct {
	answer    string
	expiresAt time.Time
}

func (s *service) Captcha(context.Context) (CaptchaChallenge, error) {
	if !s.cfg.CaptchaEnabled {
		return CaptchaChallenge{Enabled: false}, nil
	}

	left, err := randomInt(1, 9)
	if err != nil {
		return CaptchaChallenge{}, err
	}
	right, err := randomInt(1, 9)
	if err != nil {
		return CaptchaChallenge{}, err
	}
	captchaID, err := randomCaptchaID()
	if err != nil {
		return CaptchaChallenge{}, err
	}
	question := fmt.Sprintf("%d + %d", left, right)
	answer := strconv.FormatInt(left+right, 10)
	expiresAt := s.now().Add(s.cfg.CaptchaTTL)

	s.captchaMu.Lock()
	s.cleanupCaptchasLocked(s.now())
	s.captchaChallenges[captchaID] = captchaState{answer: answer, expiresAt: expiresAt}
	s.captchaMu.Unlock()

	return CaptchaChallenge{
		CaptchaID: captchaID,
		Enabled:   true,
		ExpiresAt: expiresAt,
		Image:     captchaImage(question),
	}, nil
}

func (s *service) validateLoginCaptcha(captchaID string, captchaCode string) error {
	if !s.cfg.CaptchaEnabled {
		return nil
	}

	captchaID = strings.TrimSpace(captchaID)
	captchaCode = strings.TrimSpace(captchaCode)
	if captchaID == "" || captchaCode == "" {
		return ErrCaptchaRequired
	}

	now := s.now()
	s.captchaMu.Lock()
	defer s.captchaMu.Unlock()
	s.cleanupCaptchasLocked(now)
	state, ok := s.captchaChallenges[captchaID]
	delete(s.captchaChallenges, captchaID)
	if !ok || now.After(state.expiresAt) {
		return ErrCaptchaInvalid
	}
	if captchaCode != state.answer {
		return ErrCaptchaInvalid
	}
	return nil
}

func (s *service) cleanupCaptchasLocked(now time.Time) {
	for id, state := range s.captchaChallenges {
		if !now.After(state.expiresAt) {
			continue
		}
		delete(s.captchaChallenges, id)
	}
}

func randomInt(minValue int64, maxValue int64) (int64, error) {
	if maxValue < minValue {
		return 0, ErrInvalidInput
	}
	value, err := rand.Int(rand.Reader, big.NewInt(maxValue-minValue+1))
	if err != nil {
		return 0, err
	}
	return value.Int64() + minValue, nil
}

func randomCaptchaID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func captchaImage(question string) string {
	escaped := html.EscapeString(question)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="160" height="52" viewBox="0 0 160 52">
<rect width="160" height="52" rx="8" fill="#eef6ff"/>
<path d="M8 40 C32 8, 58 50, 86 18 S132 40, 152 14" fill="none" stroke="#9cc3e6" stroke-width="2" opacity=".85"/>
<path d="M14 16 L144 44 M18 42 L148 12" stroke="#c8d6e5" stroke-width="1" opacity=".7"/>
<text x="80" y="34" text-anchor="middle" font-family="ui-monospace, SFMono-Regular, Consolas, monospace" font-size="22" font-weight="700" fill="#1f6feb" transform="rotate(-4 80 28)">%s = ?</text>
</svg>`, escaped)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
