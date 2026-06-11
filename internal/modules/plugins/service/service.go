package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/pkg/configloader"
	"github.com/rei0721/go-scaffold/pkg/logger"
)

const (
	HeaderPluginID           = "X-Aoi-Plugin-ID"
	HeaderUserID             = "X-Aoi-User-ID"
	HeaderOrgID              = "X-Aoi-Org-ID"
	HeaderTraceID            = "X-Aoi-Trace-ID"
	HeaderSignature          = "X-Aoi-Signature"
	HeaderSignatureTimestamp = "X-Aoi-Signature-Timestamp"
)

var (
	ErrDisabled       = errors.New("plugins disabled")
	ErrPluginNotFound = errors.New("plugin not found")
	ErrProxyForbidden = errors.New("plugin proxy path forbidden")
	ErrSecretMissing  = errors.New("plugin secret missing")
)

type Config struct {
	Enabled       bool
	ManifestPaths []string
	Inline        []Manifest
	HealthTimeout time.Duration
	ProxyTimeout  time.Duration
}

type Manifest struct {
	ID          string       `json:"id" yaml:"id"`
	Name        string       `json:"name" yaml:"name"`
	Version     string       `json:"version" yaml:"version"`
	BaseURL     string       `json:"baseURL" yaml:"baseURL"`
	HealthPath  string       `json:"healthPath" yaml:"healthPath"`
	Frontend    Frontend     `json:"frontend" yaml:"frontend"`
	Menus       []Menu       `json:"menus" yaml:"menus"`
	Permissions []Permission `json:"permissions" yaml:"permissions"`
	Proxy       Proxy        `json:"proxy" yaml:"proxy"`
	SecretRef   string       `json:"secretRef" yaml:"secretRef"`
}

type Frontend struct {
	Entry string `json:"entry" yaml:"entry"`
}

type Menu struct {
	Code       string `json:"code" yaml:"code"`
	Label      string `json:"label" yaml:"label"`
	Icon       string `json:"icon,omitempty" yaml:"icon"`
	Path       string `json:"path" yaml:"path"`
	Permission string `json:"permission,omitempty" yaml:"permission"`
	Order      int    `json:"order,omitempty" yaml:"order"`
}

type Permission struct {
	Code        string `json:"code" yaml:"code"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description"`
}

type Proxy struct {
	Prefixes []string `json:"prefixes" yaml:"prefixes"`
}

type HealthStatus struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	DurationMS int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

type ProxyIdentity struct {
	UserID  string
	OrgID   string
	TraceID string
}

type ProxyRequest struct {
	PluginID string
	Method   string
	Path     string
	RawQuery string
	Headers  http.Header
	Body     io.Reader
	Identity ProxyIdentity
}

type ProxyResponse struct {
	StatusCode  int
	ContentType string
	Headers     http.Header
	Body        []byte
}

type Service interface {
	List(context.Context) ([]Manifest, error)
	Get(context.Context, string) (Manifest, error)
	Health(context.Context, string) (HealthStatus, error)
	Proxy(context.Context, ProxyRequest) (ProxyResponse, error)
}

type service struct {
	enabled      bool
	plugins      map[string]Manifest
	healthClient *http.Client
	proxyClient  *http.Client
	log          logger.Logger
}

func New(cfg Config, log logger.Logger) (Service, error) {
	if cfg.HealthTimeout <= 0 {
		cfg.HealthTimeout = 3 * time.Second
	}
	if cfg.ProxyTimeout <= 0 {
		cfg.ProxyTimeout = 30 * time.Second
	}
	if !cfg.Enabled {
		return &service{enabled: false, plugins: map[string]Manifest{}, log: log}, nil
	}

	plugins := make(map[string]Manifest)
	for _, manifest := range cfg.Inline {
		if err := registerManifest(plugins, manifest); err != nil {
			return nil, err
		}
	}
	for _, path := range cfg.ManifestPaths {
		manifest, err := loadManifestFile(path)
		if err != nil {
			return nil, err
		}
		if err := registerManifest(plugins, manifest); err != nil {
			return nil, err
		}
	}

	return &service{
		enabled:      true,
		plugins:      plugins,
		healthClient: &http.Client{Timeout: cfg.HealthTimeout},
		proxyClient:  &http.Client{Timeout: cfg.ProxyTimeout},
		log:          log,
	}, nil
}

func (s *service) List(context.Context) ([]Manifest, error) {
	if !s.enabled {
		return nil, ErrDisabled
	}
	items := make([]Manifest, 0, len(s.plugins))
	for _, plugin := range s.plugins {
		items = append(items, cloneManifest(plugin))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (s *service) Get(_ context.Context, id string) (Manifest, error) {
	if !s.enabled {
		return Manifest{}, ErrDisabled
	}
	plugin, ok := s.plugins[strings.TrimSpace(id)]
	if !ok {
		return Manifest{}, ErrPluginNotFound
	}
	return cloneManifest(plugin), nil
}

func (s *service) Health(ctx context.Context, id string) (HealthStatus, error) {
	plugin, err := s.Get(ctx, id)
	if err != nil {
		return HealthStatus{}, err
	}
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(plugin.BaseURL, plugin.HealthPath, ""), nil)
	if err != nil {
		return HealthStatus{}, err
	}
	resp, err := s.healthClient.Do(req)
	status := HealthStatus{ID: plugin.ID, Status: "ok", DurationMS: time.Since(started).Milliseconds()}
	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		return status, nil
	}
	defer resp.Body.Close()
	status.StatusCode = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		status.Status = "unhealthy"
	}
	return status, nil
}

func (s *service) Proxy(ctx context.Context, req ProxyRequest) (ProxyResponse, error) {
	plugin, err := s.Get(ctx, req.PluginID)
	if err != nil {
		return ProxyResponse{}, err
	}
	path := normalizePath(req.Path, "")
	if path == "" || !proxyPathAllowed(path, plugin.Proxy.Prefixes) {
		return ProxyResponse{}, ErrProxyForbidden
	}
	secret := resolveSecret(plugin.SecretRef)
	if secret == "" {
		return ProxyResponse{}, ErrSecretMissing
	}

	body, err := readProxyBody(req.Body)
	if err != nil {
		return ProxyResponse{}, err
	}
	targetReq, err := http.NewRequestWithContext(ctx, req.Method, joinURL(plugin.BaseURL, path, req.RawQuery), bytes.NewReader(body))
	if err != nil {
		return ProxyResponse{}, err
	}
	copyProxyRequestHeaders(targetReq.Header, req.Headers)
	signProxyRequest(targetReq.Header, plugin.ID, req.Identity, req.Method, path, secret)

	resp, err := s.proxyClient.Do(targetReq)
	if err != nil {
		return ProxyResponse{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProxyResponse{}, err
	}
	return ProxyResponse{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Headers:     filteredResponseHeaders(resp.Header),
		Body:        raw,
	}, nil
}

func loadManifestFile(path string) (Manifest, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Manifest{}, fmt.Errorf("plugin manifest path is required")
	}
	var manifest Manifest
	loader := configloader.New()
	loader.SetConfigFile(path)
	if err := loader.ReadInConfig(); err != nil {
		return Manifest{}, fmt.Errorf("read plugin manifest %s: %w", path, err)
	}
	if err := loader.Unmarshal(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse plugin manifest %s: %w", path, err)
	}
	return manifest, nil
}

func registerManifest(plugins map[string]Manifest, manifest Manifest) error {
	if err := validateManifest(&manifest); err != nil {
		return err
	}
	if _, ok := plugins[manifest.ID]; ok {
		return fmt.Errorf("duplicate plugin id %q", manifest.ID)
	}
	plugins[manifest.ID] = manifest
	return nil
}

func validateManifest(manifest *Manifest) error {
	manifest.ID = strings.TrimSpace(manifest.ID)
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Version = strings.TrimSpace(manifest.Version)
	manifest.BaseURL = strings.TrimRight(strings.TrimSpace(manifest.BaseURL), "/")
	manifest.HealthPath = normalizePath(manifest.HealthPath, "/health")
	manifest.Frontend.Entry = strings.TrimSpace(manifest.Frontend.Entry)
	manifest.SecretRef = strings.TrimSpace(manifest.SecretRef)
	if manifest.ID == "" || manifest.Name == "" || manifest.Version == "" || manifest.BaseURL == "" {
		return fmt.Errorf("plugin manifest requires id, name, version, and baseURL")
	}
	for i := range manifest.Menus {
		manifest.Menus[i].Code = strings.TrimSpace(manifest.Menus[i].Code)
		manifest.Menus[i].Label = strings.TrimSpace(manifest.Menus[i].Label)
		manifest.Menus[i].Path = normalizePath(manifest.Menus[i].Path, "/")
		manifest.Menus[i].Permission = strings.TrimSpace(manifest.Menus[i].Permission)
		if manifest.Menus[i].Code == "" || manifest.Menus[i].Label == "" {
			return fmt.Errorf("plugin %s menu requires code and label", manifest.ID)
		}
	}
	for i := range manifest.Permissions {
		manifest.Permissions[i].Code = strings.TrimSpace(manifest.Permissions[i].Code)
		manifest.Permissions[i].Name = strings.TrimSpace(manifest.Permissions[i].Name)
		if manifest.Permissions[i].Code == "" {
			return fmt.Errorf("plugin %s permission requires code", manifest.ID)
		}
		if manifest.Permissions[i].Name == "" {
			manifest.Permissions[i].Name = manifest.Permissions[i].Code
		}
	}
	for i := range manifest.Proxy.Prefixes {
		manifest.Proxy.Prefixes[i] = normalizePath(manifest.Proxy.Prefixes[i], "")
		if manifest.Proxy.Prefixes[i] == "" {
			return fmt.Errorf("plugin %s proxy prefix is empty", manifest.ID)
		}
	}
	if len(manifest.Proxy.Prefixes) > 0 && manifest.SecretRef == "" {
		return fmt.Errorf("plugin %s secretRef is required for proxy", manifest.ID)
	}
	return nil
}

func normalizePath(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	value = "/" + strings.Trim(strings.TrimRight(value, "/"), "/")
	if value == "/" && fallback != "" {
		return value
	}
	return value
}

func joinURL(baseURL, path, rawQuery string) string {
	target := strings.TrimRight(baseURL, "/") + normalizePath(path, "/")
	if rawQuery != "" {
		target += "?" + rawQuery
	}
	return target
}

func proxyPathAllowed(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = normalizePath(prefix, "")
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func resolveSecret(secretRef string) string {
	if secretRef == "" {
		return ""
	}
	if value := os.Getenv(secretRef); value != "" {
		return value
	}
	return ""
}

func readProxyBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return io.ReadAll(body)
}

func copyProxyRequestHeaders(dst, src http.Header) {
	for key, values := range src {
		if skipRequestHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func skipRequestHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "cookie", "host", "content-length", "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func signProxyRequest(headers http.Header, pluginID string, identity ProxyIdentity, method, path, secret string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := strings.Join([]string{
		strings.ToUpper(method),
		path,
		identity.OrgID,
		identity.UserID,
		identity.TraceID,
		timestamp,
	}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))

	headers.Set(HeaderPluginID, pluginID)
	headers.Set(HeaderUserID, identity.UserID)
	headers.Set(HeaderOrgID, identity.OrgID)
	headers.Set(HeaderTraceID, identity.TraceID)
	headers.Set(HeaderSignatureTimestamp, timestamp)
	headers.Set(HeaderSignature, hex.EncodeToString(mac.Sum(nil)))
}

func filteredResponseHeaders(src http.Header) http.Header {
	out := http.Header{}
	for key, values := range src {
		if skipResponseHeader(key) {
			continue
		}
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}

func skipResponseHeader(key string) bool {
	switch strings.ToLower(key) {
	case "content-length", "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func cloneManifest(src Manifest) Manifest {
	dst := src
	dst.Menus = append([]Menu(nil), src.Menus...)
	dst.Permissions = append([]Permission(nil), src.Permissions...)
	dst.Proxy.Prefixes = append([]string(nil), src.Proxy.Prefixes...)
	sort.Slice(dst.Menus, func(i, j int) bool {
		if dst.Menus[i].Order == dst.Menus[j].Order {
			return dst.Menus[i].Code < dst.Menus[j].Code
		}
		return dst.Menus[i].Order < dst.Menus[j].Order
	})
	return dst
}
