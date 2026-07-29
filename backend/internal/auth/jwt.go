package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/config"
)

// TokenPair represents an access + refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
	Mobile string    `json:"mobile"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	cfg        config.JWTConfig
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	m := &JWTManager{cfg: cfg}

	// Load or generate RSA keys
	if err := m.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load JWT keys: %w", err)
	}

	return m, nil
}

func (m *JWTManager) loadKeys() error {
	privateBytes, err := os.ReadFile(m.cfg.PrivateKeyPath)
	if err == nil {
		// Keys exist, parse them
		block, _ := pem.Decode(privateBytes)
		if block == nil {
			return fmt.Errorf("failed to decode private key PEM")
		}
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// Try PKCS1
			m.privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse private key: %w", err)
			}
		} else {
			var ok bool
			m.privateKey, ok = parsedKey.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("private key is not RSA")
			}
		}

		publicBytes, err := os.ReadFile(m.cfg.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}
		block, _ = pem.Decode(publicBytes)
		if block == nil {
			return fmt.Errorf("failed to decode public key PEM")
		}
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %w", err)
		}
		m.publicKey = pubKey.(*rsa.PublicKey)

		return nil
	}

	// Keys do not exist — generate new pair
	return m.generateKeys()
}

func (m *JWTManager) generateKeys() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}
	m.privateKey = key
	m.publicKey = &key.PublicKey

	// Ensure key directory exists
	os.MkdirAll("keys", 0700)

	// Save private key
	privBytes, _ := x509.MarshalPKCS8PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(m.cfg.PrivateKeyPath, privPEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile(m.cfg.PublicKeyPath, pubPEM, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

// GenerateTokenPair creates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, role, mobile string) (*TokenPair, error) {
	now := time.Now()
	tokenID := uuid.New().String()

	// Access Token
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID.String(),
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.AccessTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID: userID,
		Role:   role,
		Mobile: mobile,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token (with longer expiry)
	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID.String(),
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.RefreshTTL)),
		},
		UserID: userID,
		Role:   role,
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(m.cfg.AccessTTL.Seconds()),
	}, nil
}

// ValidateToken parses and validates a JWT token
func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	}, jwt.WithIssuer(m.cfg.Issuer), jwt.WithLeeway(30*time.Second))

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
