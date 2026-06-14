package ports

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"
)

type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
	Sync() error
}

type IDGenerator interface {
	NextID() int64
	NextIDString() string
}

type I18n interface {
	T(string, string, ...map[string]interface{}) string
	MustT(string, string, ...map[string]interface{}) string
	IsSupported(string) bool
	GetDefaultLanguage() string
	LoadMessages(string) error
}

type HTTPContext interface {
	Request() *http.Request
	RequestContext() context.Context
	GetHeader(string) string
	Header(string, string)
	Set(string, any)
	Get(any) (any, bool)
	Param(string) string
	BindJSON(any) error
	JSON(int, any)
	Data(int, string, []byte)
	AbortWithStatusJSON(int, any)
	Next()
	Path() string
	Method() string
	ClientIP() string
	Status() int
}

type HTTPHandlerFunc func(HTTPContext)

type HTTPRouter interface {
	Use(...HTTPHandlerFunc)
	GET(string, HTTPHandlerFunc)
	POST(string, HTTPHandlerFunc)
	PATCH(string, HTTPHandlerFunc)
	PUT(string, HTTPHandlerFunc)
	DELETE(string, HTTPHandlerFunc)
	ANY(string, HTTPHandlerFunc)
	Group(string) HTTPRouter
}

type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type RouteLister interface {
	Routes() []RouteInfo
}

type StaticSPAConfig struct {
	MountPath string
	DistDir   string
}

var ErrStaticSPAIndexMissing = errors.New("static spa index.html missing")

type StaticSPAMounter interface {
	MountStaticSPA(StaticSPAConfig) error
}

type CORSConfig struct {
	Enabled          bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

type CORSFactory func(CORSConfig) HTTPHandlerFunc

type MediaObjectStorage interface {
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, os.FileMode) error
	Remove(string) error
	RemoveAll(string) error
	MkdirAll(string, os.FileMode) error
	DetectMIMEFromBytes([]byte) (string, error)
}

type HostMetricsCollector interface {
	Collect(context.Context) HostMetrics
}

type HostMetrics struct {
	CPU  CPUInfo
	RAM  RAMInfo
	Disk []DiskInfo
}

type CPUInfo struct {
	Cores   int
	Percent []float64
}

type RAMInfo struct {
	TotalMB     uint64
	UsedMB      uint64
	UsedPercent float64
}

type DiskInfo struct {
	FSType      string
	MountPoint  string
	TotalGB     uint64
	TotalMB     uint64
	UsedGB      uint64
	UsedMB      uint64
	UsedPercent float64
}

type PasswordCrypto interface {
	HashPassword(string) (string, error)
	VerifyPassword(string, string) error
}

type TokenSubject struct {
	UserID    int64
	OrgID     int64
	SessionID int64
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type TokenClaims struct {
	UserID    int64
	OrgID     int64
	SessionID int64
	TokenType string
}

type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshTokenHash string
	RefreshExpiresAt time.Time
}

type TokenManager interface {
	IssueAccess(context.Context, TokenSubject) (string, time.Time, error)
	IssueRefresh(context.Context) (string, string, time.Time, error)
	IssuePair(context.Context, TokenSubject) (TokenPair, error)
	Parse(context.Context, string, string) (*TokenClaims, error)
	HashRefreshToken(string) string
}

type AuthorizationRule struct {
	PType  string
	Values []string
}

type AuthorizerEnforcer interface {
	Enforce(context.Context, string, string, string, string) (bool, error)
	AddPolicy(context.Context, string, string, string, string) (bool, error)
	AddRoleForUser(context.Context, string, string, string) (bool, error)
	DeleteRoleForUser(context.Context, string, string, string) (bool, error)
	GetRolesForUser(context.Context, string, string) ([]string, error)
	LoadRules(context.Context, []AuthorizationRule) error
}

type TOTPKey struct {
	Secret string
	URL    string
}

type TOTPProvider interface {
	GenerateTOTP(string, string) (TOTPKey, error)
	ValidateTOTP(string, string) bool
}

type RPCHandlerFunc func(context.Context, json.RawMessage) (any, error)

type RPCRegistry interface {
	Register(string, RPCHandlerFunc) error
	Methods() []string
}

type RPCError struct {
	Code    int
	Message string
	Data    any
}

func (e *RPCError) Error() string {
	return e.Message
}

func InvalidRPCParams(message string) *RPCError {
	return &RPCError{Code: -32602, Message: message}
}

type ManifestLoader interface {
	LoadManifestFile(string, any) error
}

type ProxyHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ReadCloserFactory func(io.Reader) io.ReadCloser
