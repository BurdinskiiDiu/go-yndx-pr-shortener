package authentication

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var key = []byte("secretKey")

func randID() ([]byte, error) {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.New("creating new unique user ID error, " + err.Error())
	}
	return b, nil
}

func CreateUserID() (string, string, error) {
	id, err := randID()
	idStr := hex.EncodeToString(id)
	if err != nil {
		return "nil", "", err
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte(idStr))
	signature := h.Sum(nil)
	signatureStr := hex.EncodeToString(signature)
	return idStr, signatureStr, nil
}

func CheckCookie(cookieStr string) (string, string, error) {
	if cookieStr == "" {
		return "", "", errors.New("cookie is empty")
	}
	if !(len([]byte(cookieStr)) > 8) {
		return "", "", errors.New("wrong cookie len")
	}
	decGotStr, err := hex.DecodeString(cookieStr)
	if err != nil {
		return "", "", errors.New("decoding gotted cookie string err, " + err.Error())
	}
	gottedUserID := decGotStr[:8]
	gottedUserIDStr := hex.EncodeToString(gottedUserID)
	gottedSign := decGotStr[8:]
	gottedSignStr := hex.EncodeToString(gottedSign)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(gottedUserIDStr))
	signature := h.Sum(nil)
	decSign := hex.EncodeToString(signature)

	if hmac.Equal([]byte(decSign), []byte(gottedSignStr)) {
		return gottedUserIDStr, gottedSignStr, nil
	}
	return "", "", errors.New("wrong signature")
}
