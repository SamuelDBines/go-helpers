package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Jwt struct {
	Secret []byte `json:"secret"`
}

type Claims struct {
	UserID int64  `json:"uid"`
	Sub    string `json:"sub"`
	Iss    string `json:"iss,omitempty"`
	Aud    string `json:"aud,omitempty"`
	Exp    int64  `json:"exp,omitempty"`
	Iat    int64  `json:"iat,omitempty"`
}

func base64urlEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func signHS256(message string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	return base64urlEncode(signature)
}

func CreateJWT(claims Claims, secret []byte) (string, error) {
	header := Header{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := base64urlEncode(headerJSON)
	payloadPart := base64urlEncode(payloadJSON)

	unsignedToken := headerPart + "." + payloadPart
	signaturePart := signHS256(unsignedToken, secret)

	return unsignedToken + "." + signaturePart, nil
}

func VerifyJWT(token string, secret []byte) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerPart := parts[0]
	payloadPart := parts[1]
	signaturePart := parts[2]

	unsignedToken := headerPart + "." + payloadPart

	expectedSignature := signHS256(unsignedToken, secret)

	if !hmac.Equal([]byte(signaturePart), []byte(expectedSignature)) {
		return nil, errors.New("invalid token signature")
	}

	headerBytes, err := base64urlDecode(headerPart)
	if err != nil {
		return nil, errors.New("invalid header encoding")
	}

	var header Header
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errors.New("invalid header json")
	}

	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, errors.New("unsupported token header")
	}

	payloadBytes, err := base64urlDecode(payloadPart)
	if err != nil {
		return nil, errors.New("invalid payload encoding")
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("invalid payload json")
	}

	if claims.Exp != 0 && time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func (jwt *Jwt) CreateJWT(claims Claims) (string, error) {
	return CreateJWT(claims, jwt.Secret)
}

func (jwt *Jwt) VerifyJWT(token string) (*Claims, error) {

	return VerifyJWT(token, jwt.Secret)
}

func (jwt *Jwt) SignUser(userID int64, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}
	return jwt.CreateJWT(claims)
}

func (jwt *Jwt) ParseUser(token string) (int64, error) {
	claims, err := jwt.VerifyJWT(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
