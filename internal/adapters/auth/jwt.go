package authadapter

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/domain"
)

// NewVerifier returns a JWT verifier for the config, or nil when disabled.
func NewVerifier(cfg appauth.Config) (appauth.TokenVerifier, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	if cfg.Audience == "" {
		return nil, fmt.Errorf("MEMLORE_OIDC_AUDIENCE is required when OIDC is enabled")
	}
	if cfg.HMACSecret != "" {
		return &HMACVerifier{Config: cfg}, nil
	}
	return &JWKSVerifier{
		Config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// HMACVerifier validates HS256 tokens with a shared secret (tests / simple deploys).
type HMACVerifier struct {
	Config appauth.Config
}

func (v *HMACVerifier) Verify(_ context.Context, rawToken string) (domain.Principal, error) {
	return parseAndValidate(rawToken, v.Config, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, &domain.UnauthorizedError{Message: "unexpected signing method"}
		}
		return []byte(v.Config.HMACSecret), nil
	})
}

// JWKSVerifier validates RS256 tokens using a remote JWKS URL.
type JWKSVerifier struct {
	Config appauth.Config
	client *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func (v *JWKSVerifier) Verify(ctx context.Context, rawToken string) (domain.Principal, error) {
	return parseAndValidate(rawToken, v.Config, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, &domain.UnauthorizedError{Message: "unexpected signing method"}
		}
		kid, _ := token.Header["kid"].(string)
		return v.lookupKey(ctx, kid)
	})
}

func (v *JWKSVerifier) lookupKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.keys == nil || time.Since(v.fetchedAt) > 5*time.Minute {
		if err := v.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	if kid != "" {
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
		return nil, &domain.UnauthorizedError{Message: "unknown token key id"}
	}
	for _, key := range v.keys {
		return key, nil
	}
	return nil, &domain.UnauthorizedError{Message: "no jwks keys available"}
}

func (v *JWKSVerifier) refreshLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.Config.JWKSURL, nil)
	if err != nil {
		return &domain.UnauthorizedError{Message: "invalid jwks url"}
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return &domain.UnauthorizedError{Message: "jwks fetch failed"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &domain.UnauthorizedError{Message: "jwks fetch failed"}
	}
	var doc struct {
		Keys []jwkKey `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return &domain.UnauthorizedError{Message: "invalid jwks document"}
	}
	keys := map[string]*rsa.PublicKey{}
	for i, k := range doc.Keys {
		if strings.ToUpper(k.Kty) != "RSA" {
			continue
		}
		pub, err := k.rsaPublicKey()
		if err != nil {
			continue
		}
		kid := k.Kid
		if kid == "" {
			kid = fmt.Sprintf("key-%d", i)
		}
		keys[kid] = pub
	}
	if len(keys) == 0 {
		return &domain.UnauthorizedError{Message: "no usable jwks keys"}
	}
	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (k jwkKey) rsaPublicKey() (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	if len(eb) == 0 {
		return nil, fmt.Errorf("empty exponent")
	}
	var e int
	for _, b := range eb {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}, nil
}

func parseAndValidate(rawToken string, cfg appauth.Config, keyFunc jwt.Keyfunc) (domain.Principal, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg(), jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithAudience(cfg.Audience),
		jwt.WithExpirationRequired(),
	)
	token, err := parser.Parse(rawToken, keyFunc)
	if err != nil || !token.Valid {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "invalid token"}
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "invalid token claims"}
	}
	sub, _ := claims["sub"].(string)
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "token subject missing"}
	}
	role, err := roleFromClaims(claims, cfg.RoleClaim)
	if err != nil {
		return domain.Principal{}, err
	}
	return domain.Principal{Subject: sub, Role: role}, nil
}

func roleFromClaims(claims jwt.MapClaims, claimName string) (domain.Role, error) {
	raw, ok := claims[claimName]
	if !ok {
		return "", &domain.ForbiddenError{Message: "role claim missing"}
	}
	var roles []string
	switch v := raw.(type) {
	case string:
		roles = []string{v}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				roles = append(roles, s)
			}
		}
	default:
		return "", &domain.ForbiddenError{Message: "role claim invalid"}
	}
	role, ok := domain.HighestRole(roles)
	if !ok {
		return "", &domain.ForbiddenError{Message: "unrecognized role"}
	}
	return role, nil
}

// IssueHMACToken is a test helper to mint HS256 tokens.
func IssueHMACToken(cfg appauth.Config, subject string, role domain.Role, ttl time.Duration) (string, error) {
	if cfg.HMACSecret == "" {
		return "", fmt.Errorf("hmac secret required")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":          subject,
		"iss":          cfg.Issuer,
		"aud":          cfg.Audience,
		"iat":          now.Unix(),
		"exp":          now.Add(ttl).Unix(),
		cfg.RoleClaim:  string(role),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.HMACSecret))
}
