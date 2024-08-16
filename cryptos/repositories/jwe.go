package repositories

import (
  "crypto/rsa"
  "os"
  "path"

  "github.com/go-jose/go-jose/v4"
  "golang.org/x/crypto/ssh"
)

type JweRepository struct {
  privateKey *rsa.PrivateKey
  publicKey  *rsa.PublicKey
}

func (r *JweRepository) PrivateKey() *rsa.PrivateKey {
  if r.privateKey == nil {
    home, err := os.UserHomeDir()
    if err != nil {
      panic(err)
    }
    bytes, err := os.ReadFile(path.Join(home, ".ssh/jwe_rsa"))
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

func (r *JweRepository) PublicKey() *rsa.PublicKey {
  if r.publicKey == nil {
    home, err := os.UserHomeDir()
    if err != nil {
      panic(err)
    }
    bytes, err := os.ReadFile(path.Join(home, ".ssh/client_rsa"))
    if err != nil {
      panic(err)
    }
    privateKey, err := ssh.ParseRawPrivateKey(bytes)
    if err != nil {
      panic(err)
    }
    r.publicKey = &privateKey.(*rsa.PrivateKey).PublicKey
  }
  return r.publicKey
}

func (r *JweRepository) Encrypt(payload []byte) (jweCompact string, err error) {
  enc, err := jose.NewEncrypter(
    jose.A256GCM,
    jose.Recipient{
      Algorithm: jose.RSA_OAEP_256,
      Key:       r.PublicKey(),
    },
    nil,
  )
  if err != nil {
    return
  }
  jwe, err := enc.Encrypt(payload)
  if err != nil {
    return
  }

  jweCompact, err = jwe.CompactSerialize()
  if err != nil {
    return
  }

  return
}

func (r *JweRepository) Decrypt(jweCompact string) (payload []byte, err error) {
  jwe, err := jose.ParseEncrypted(
    jweCompact,
    []jose.KeyAlgorithm{jose.RSA_OAEP_256},
    []jose.ContentEncryption{jose.A256GCM},
  )
  if err != nil {
    return
  }
  payload, err = jwe.Decrypt(r.PrivateKey())
  return
}
