package token

import (
	"encoding/json"
	"time"

	"github.com/knq/jwt"
)

// DefaultExp is token default expiration time.
const DefaultExp = time.Minute * 15

// Claims contains the registered JWT claims.
type Claims struct {

	// UID ("uid") identifies user that requested this token.
	UID string `json:"uid"`

	// Expiration ("exp") identifies the expiration time on or after which the
	// JWT MUST NOT be accepted for processing.
	Expiration json.Number `json:"exp,omitempty"`

	// IssuedAt ("iat") identifies the time at which the JWT was issued.
	IssuedAt json.Number `json:"iat,omitempty"`
}

// New create new jwt token signed using ECDSA with the P-384
// curve and the SHA-384 hash function.
func New(claims *Claims, privateKey, publicKey string) (string, error) {
	es384, err := jwt.ES384.New(jwt.PEM{
		[]byte(privateKey), []byte(publicKey),
	})
	if err != nil {
		return "", err
	}

	tokenBuf, err := es384.Encode(claims)
	if err != nil {
		return "", err
	}
	return string(tokenBuf), nil
}
