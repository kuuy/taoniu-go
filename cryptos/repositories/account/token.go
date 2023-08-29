package account

import (
  "bytes"
  "crypto/rsa"
  "errors"
  "io/ioutil"
  "os"
  "path"
  "time"

  "github.com/lestrrat/go-jwx/jwa"
  "github.com/lestrrat/go-jwx/jwt"
  "golang.org/x/crypto/ssh"
)

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
    bytes, err := ioutil.ReadFile(path.Join(home, ".ssh/jwt_rsa"))
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
  token.Set("exp", now.AddDate(0, 0, 14).Unix())

  refreshToken, err := token.Sign(jwa.RS256, r.PrivateKey())
  if err != nil {
    return "", err
  }

  return string(refreshToken), nil
}

func (r *TokenRepository) Uid(tokenString string) (string, error) {
  now := time.Now().UTC()

  token, err := jwt.Parse(
    bytes.NewReader([]byte(tokenString)),
    jwt.WithVerify(
      jwa.RS256,
      &r.PrivateKey().PublicKey,
    ),
  )
  if err != nil {
    return "", err
  }

  uid, _ := token.Get("uid")
  exp, _ := token.Get("exp")
  if now.Unix() > exp.(*jwt.NumericDate).Unix() {
    return uid.(string), errors.New("token has been expired")
  }

  return uid.(string), nil
}
