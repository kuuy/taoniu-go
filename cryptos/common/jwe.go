package common

import (
  "crypto/rsa"
  "os"
  "path"
  "sync"

  "github.com/go-jose/go-jose/v4"
  "golang.org/x/crypto/ssh"
)

var (
  globalPrivateKey     *rsa.PrivateKey
  globalPrivateKeyOnce sync.Once
  globalPublicKey      *rsa.PublicKey
  globalPublicKeyOnce  sync.Once
)

type Jwe struct {
  privateKey *rsa.PrivateKey
  publicKey  *rsa.PublicKey
}

func (r *Jwe) PrivateKey() *rsa.PrivateKey {
  if r.privateKey == nil {
    globalPrivateKeyOnce.Do(func() {
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
      globalPrivateKey = privateKey.(*rsa.PrivateKey)
    })
    r.privateKey = globalPrivateKey
  }
  return r.privateKey
}

func (r *Jwe) PublicKey() *rsa.PublicKey {
  if r.publicKey == nil {
    globalPublicKeyOnce.Do(func() {
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
      globalPublicKey = &privateKey.(*rsa.PrivateKey).PublicKey
    })
    r.publicKey = globalPublicKey
  }
  return r.publicKey
}

func (r *Jwe) Encrypt(payload []byte) (jweCompact string, err error) {
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

func (r *Jwe) Decrypt(jweCompact string) (payload []byte, err error) {
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
