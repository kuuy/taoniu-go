package account

import (
  "crypto/rsa"
  "errors"
  "log"
  "os"
  "path"
  "taoniu.local/cryptos/common"
  "time"

  "github.com/go-jose/go-jose/v4"
  "github.com/go-jose/go-jose/v4/jwt"
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
  enc, err := jose.NewEncrypter(
    jose.A128GCM,
    jose.Recipient{
      Algorithm: jose.DIRECT,
      Key:       []byte(common.GetEnvString("JWT_KEY")),
    },
    (&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"),
  )
  if err != nil {
    return
  }
  sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: r.PrivateKey()}, nil)
  if err != nil {
    return
  }

  now := time.Now().UTC()
  cl := &jwt.Claims{
    Expiry:   jwt.NewNumericDate(now.Add(15 * time.Minute)),
    IssuedAt: jwt.NewNumericDate(now),
  }
  privateCl := &TokenInfo{
    uid,
  }
  accessToken, err = jwt.SignedAndEncrypted(sig, enc).Claims(cl).Claims(privateCl).Serialize()
  return
}

func (r *TokenRepository) RefreshToken(uid string) (refreshToken string, err error) {
  enc, err := jose.NewEncrypter(
    jose.A128GCM,
    jose.Recipient{
      Algorithm: jose.DIRECT,
      Key:       []byte(common.GetEnvString("JWT_KEY")),
    },
    (&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"),
  )
  if err != nil {
    return
  }
  sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: r.PrivateKey()}, nil)
  if err != nil {
    return
  }

  now := time.Now().UTC()
  cl := &jwt.Claims{
    Expiry:   jwt.NewNumericDate(now.AddDate(0, 0, 14)),
    IssuedAt: jwt.NewNumericDate(now),
  }
  privateCl := &TokenInfo{
    uid,
  }
  refreshToken, err = jwt.SignedAndEncrypted(sig, enc).Claims(cl).Claims(privateCl).Serialize()
  return
}

func (r *TokenRepository) Uid(tokenString string) (uid string, err error) {
  tok, err := jwt.ParseSignedAndEncrypted(
    tokenString,
    []jose.KeyAlgorithm{jose.DIRECT},
    []jose.ContentEncryption{jose.A128GCM},
    []jose.SignatureAlgorithm{jose.RS256},
  )
  if err != nil {
    log.Println("token parser created failed", err)
    return
  }

  nested, err := tok.Decrypt([]byte(common.GetEnvString("JWT_KEY")))
  if err != nil {
    return
  }

  var cl *jwt.Claims
  var info *TokenInfo
  err = nested.Claims(&r.PrivateKey().PublicKey, &cl, &info)
  if err != nil {
    return
  }

  now := time.Now().UTC()
  if now.Unix() > cl.Expiry.Time().Unix() {
    err = errors.New("token has been expired")
    return
  }

  uid = info.Uid
  return
}
