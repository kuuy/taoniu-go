package common

import (
  "crypto/sha512"
  "encoding/hex"
  "math/rand"
  "time"
)

func GeneratePassword(password string, salt string) string {
  hash := sha512.New()
  hash.Write(append([]byte(password), salt...))
  hashedPassword := hash.Sum(nil)
  return hex.EncodeToString(hashedPassword)
}

func GenerateSalt(size int) string {
  characters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
  result := make([]byte, size)
  rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
  for i := range result {
    result[i] = characters[rnd.Intn(len(characters))]
  }
  return string(result)
}

func VerifyPassword(password string, salt string, currentPassword string) bool {
  hashedPassword := GeneratePassword(password, salt)
  if hashedPassword == currentPassword {
    return true
  }

  return false
}
