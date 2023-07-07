package authentication

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/google/uuid"
)

func CreateUserID(cf config.Config) (string, string, error) {
	id := uuid.NewString()
	h := hmac.New(sha256.New, []byte(cf.AuthentKey))
	h.Write([]byte(id))
	signature := h.Sum(nil)
	var cookieSls []byte
	cookieSls = signature

	cookieSls = append(cookieSls, []byte(id)...)
	cookieStr := hex.EncodeToString(cookieSls)

	fmt.Println("created id is " + id)
	fmt.Println(len([]byte(id)))
	fmt.Println("created signature is " + string(signature))
	fmt.Println(len(signature))
	fmt.Println("created sookieSls is " + string(cookieSls))
	fmt.Println(len([]byte(cookieSls)))
	fmt.Println("created cookieStr is " + cookieStr)
	fmt.Println(len([]byte(cookieStr)))
	return id, cookieStr, nil
}

func CheckCookie(cookieStr string, cf config.Config) (string, error) {
	if cookieStr == "" {
		return "", errors.New("cookie is empty")
	}
	if !(len([]byte(cookieStr)) > 32) {
		return "", errors.New("wrong cookie len")
	}
	gCookieSlc, err := hex.DecodeString(cookieStr)

	if err != nil {
		return "", fmt.Errorf("decoding gotted cookies error: %w", err)
	}
	gottedUserID := string(gCookieSlc[32:])
	gottedSign := gCookieSlc[:32]

	h := hmac.New(sha256.New, []byte(cf.AuthentKey))
	h.Write([]byte(gottedUserID))
	signature := h.Sum(nil)

	if hmac.Equal(signature, gottedSign) {
		return gottedUserID, nil
	}
	return "", errors.New("wrong signature")
}
