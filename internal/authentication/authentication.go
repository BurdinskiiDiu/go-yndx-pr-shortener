package authentication

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	fmt.Println("created user is " + string(id))
	fmt.Println("created userStr is " + string(idStr))
	if err != nil {
		return "nil", "", err
	}
	h := hmac.New(sha256.New, key)
	h.Write(id)
	signature := h.Sum(nil)
	signatureStr := hex.EncodeToString(signature)
	fmt.Println("created signature is " + string(signature))
	fmt.Println("created signatureStr is " + string(signatureStr))
	//return string(id), string(signature), nil
	return idStr, signatureStr, nil
}

func CheckCookie(cookieStr string) (string, string, error) {
	if cookieStr == "" {
		return "", "", errors.New("cookie is empty")
	}
	if !(len([]byte(cookieStr)) > 8) {
		return "", "", errors.New("wrong cookie len")
	}
	gottedUserID := []byte(cookieStr)[:8]
	gottedSignature := []byte(cookieStr)[8:]

	h := hmac.New(sha256.New, key)
	h.Write(gottedUserID)
	signature := h.Sum(nil)
	if hmac.Equal(signature, gottedSignature) {
		return string(gottedUserID), string(gottedSignature), nil
	}
	return "", "", errors.New("wrong signature")
}
