package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Verify(secret, code string, now time.Time) bool {
	secret = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	code = strings.TrimSpace(code)
	if secret == "" || len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		expected := Generate(secret, now.Add(time.Duration(offset)*30*time.Second))
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func Generate(secret string, now time.Time) string {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return ""
	}
	counter := uint64(now.Unix() / 30)
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], counter)
	mac := hmac.New(sha1.New, decoded)
	_, _ = mac.Write(message[:])
	hash := mac.Sum(nil)
	offset := hash[len(hash)-1] & 0x0f
	value := (uint32(hash[offset])&0x7f)<<24 | uint32(hash[offset+1])<<16 | uint32(hash[offset+2])<<8 | uint32(hash[offset+3])
	return fmt.Sprintf("%06d", value%1000000)
}
