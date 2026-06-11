// Package token hides JWT and refresh-token hashing details behind project APIs.
package token

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidConfig = errors.New("invalid token config")
	ErrInvalidToken  = errors.New("invalid token")
	ErrWrongType     = errors.New("wrong token type")
)

type Config struct {
	Issuer        string
	Audience      []string
	SigningKey    string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	RefreshPepper string
	Now           func() time.Time
}

type Subject struct {
	UserID    int64
	OrgID     int64
	SessionID int64
}

type Claims struct {
	UserID    int64  `json:"userId"`
	OrgID     int64  `json:"orgId"`
	SessionID int64  `json:"sessionId"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

type Pair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshTokenHash string
	RefreshExpiresAt time.Time
}

type Manager interface {
	IssueAccess(context.Context, Subject) (string, time.Time, error)
	IssueRefresh(context.Context) (string, string, time.Time, error)
	IssuePair(context.Context, Subject) (Pair, error)
	Parse(context.Context, string, string) (*Claims, error)
	HashRefreshToken(string) string
}

type manager struct {
	cfg Config
}

func New(cfg Config) (Manager, error) {
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("%w: issuer is required", ErrInvalidConfig)
	}
	if cfg.SigningKey == "" {
		return nil, fmt.Errorf("%w: signing key is required", ErrInvalidConfig)
	}
	if len(cfg.SigningKey) < 32 {
		return nil, fmt.Errorf("%w: signing key must be at least 32 bytes", ErrInvalidConfig)
	}
	if cfg.AccessTTL <= 0 {
		return nil, fmt.Errorf("%w: access ttl must be positive", ErrInvalidConfig)
	}
	if cfg.RefreshTTL <= 0 {
		return nil, fmt.Errorf("%w: refresh ttl must be positive", ErrInvalidConfig)
	}
	if cfg.RefreshPepper == "" {
		return nil, fmt.Errorf("%w: refresh pepper is required", ErrInvalidConfig)
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &manager{cfg: cfg}, nil
}

func (m *manager) IssueAccess(_ context.Context, subject Subject) (string, time.Time, error) {
	now := m.cfg.Now().UTC()
	expiresAt := now.Add(m.cfg.AccessTTL)
	claims := Claims{
		UserID:    subject.UserID,
		OrgID:     subject.OrgID,
		SessionID: subject.SessionID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Audience:  jwt.ClaimStrings(m.cfg.Audience),
			Subject:   fmt.Sprintf("%d", subject.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.SigningKey))
	if err != nil {
		return "", time.Time{}, err
	}
	return raw, expiresAt, nil
}

func (m *manager) IssueRefresh(_ context.Context) (string, string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, m.HashRefreshToken(token), m.cfg.Now().UTC().Add(m.cfg.RefreshTTL), nil
}

func (m *manager) IssuePair(ctx context.Context, subject Subject) (Pair, error) {
	access, accessExpiresAt, err := m.IssueAccess(ctx, subject)
	if err != nil {
		return Pair{}, err
	}
	refresh, refreshHash, refreshExpiresAt, err := m.IssueRefresh(ctx)
	if err != nil {
		return Pair{}, err
	}
	return Pair{
		AccessToken:      access,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     refresh,
		RefreshTokenHash: refreshHash,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (m *manager) Parse(_ context.Context, raw string, expectedType string) (*Claims, error) {
	claims := &Claims{}
	opts := []jwt.ParserOption{
		jwt.WithIssuer(m.cfg.Issuer),
	}
	if len(m.cfg.Audience) > 0 {
		opts = append(opts, jwt.WithAudience(m.cfg.Audience...))
	}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(m.cfg.SigningKey), nil
	}, opts...)
	if err != nil || parsed == nil || !parsed.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrWrongType
	}
	if claims.UserID <= 0 || claims.SessionID <= 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *manager) HashRefreshToken(raw string) string {
	mac := hmac.New(sha256.New, []byte(m.cfg.RefreshPepper))
	_, _ = mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}
