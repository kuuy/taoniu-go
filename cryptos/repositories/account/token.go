package account

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/lestrrat/go-jwx/jwa"
	"github.com/lestrrat/go-jwx/jwt"
	"time"
)

type TokenRepository struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func (r *TokenRepository) PrivateKey() *rsa.PrivateKey {
	if r.privateKey == nil {
		r.privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	}
	return r.privateKey
}

func (r *TokenRepository) AccessToken(uid string) (string, error) {
	now := time.Now().UTC()

	token := jwt.New()
	token.Set("uid", uid)
	token.Set("iat", now.Unix())
	token.Set("exp", now.Add(15*time.Minute).Unix())

	accessToken, err := token.Sign(jwa.RS256, r.PrivateKey())
	if err != nil {
		return "", err
	}

	return string(accessToken), nil
}

func (r *TokenRepository) RefreshToken(uid string) (string, error) {
	now := time.Now().UTC()

	token := jwt.New()
	token.Set("uid", uid)
	token.Set("iat", now.Unix())
	token.Set("iat", now.Unix())
	token.Set("exp", now.Add(15*time.Minute).Unix())

	refreshToken, err := token.Sign(jwa.RS256, r.PrivateKey())
	if err != nil {
		return "", err
	}

	return string(refreshToken), nil
}

func (r *TokenRepository) Verify(tokenString string) error {
	return nil
}
