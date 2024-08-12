package account

import (
  "crypto/rsa"
  "errors"
  "os"
  "path"
  "time"

  "github.com/go-jose/go-jose/v4"
  "github.com/go-jose/go-jose/v4/jwt"
  "golang.org/x/crypto/ssh"
)

type TokenClaim struct {
  Uid      string `json:"uid"`
  Expiry   int64  `json:"iat"`
  IssuedAt int64  `json:"exp"`
}

type TokenRepository struct {
  privateKey *rsa.PrivateKey
  publicKey  *rsa.PublicKey
}

func (r *TokenRepository) PrivateKey() *rsa.PrivateKey {
  if r.privateKey == nil {
    home, err := os.UserHomeDir()
    if err != nil {
      panic(err)
    }
    bytes, err := os.ReadFile(path.Join(home, ".ssh/jwt_rsa"))
    if err != nil {
      panic(err)
    }
    privateKey, err := ssh.ParseRawPrivateKey(bytes)
    if err != nil {
      panic(err)
    }
    r.privateKey = privateKey.(*rsa.PrivateKey)
  }
  return r.privateKey
}

func (r *TokenRepository) AccessToken(uid string) (accessToken string, err error) {
  now := time.Now().UTC()

  sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: r.PrivateKey()}, (&jose.SignerOptions{}).WithType("JWT"))
  if err != nil {
    return
  }

  cl := &TokenClaim{
    uid,
    now.Add(15 * time.Minute).Unix(),
    now.Unix(),
  }
  accessToken, err = jwt.Signed(sig).Claims(cl).Serialize()
  return
}

func (r *TokenRepository) RefreshToken(uid string) (refreshToken string, err error) {
  now := time.Now().UTC()

  sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: r.PrivateKey()}, (&jose.SignerOptions{}).WithType("JWT"))
  if err != nil {
    return
  }

  cl := &TokenClaim{
    uid,
    now.AddDate(0, 0, 14).Unix(),
    now.Unix(),
  }
  refreshToken, err = jwt.Signed(sig).Claims(cl).Serialize()
  return
}

func (r *TokenRepository) Uid(tokenString string) (uid string, err error) {
  tok, err := jwt.ParseSigned(tokenString, []jose.SignatureAlgorithm{jose.RS256})
  if err != nil {
    return
  }

  var cl *TokenClaim
  if err = tok.UnsafeClaimsWithoutVerification(&cl); err != nil {
    return
  }

  now := time.Now().UTC()
  if now.Unix() > cl.Expiry {
    err = errors.New("token has been expired")
    return
  }

  uid = cl.Uid
  return
}
