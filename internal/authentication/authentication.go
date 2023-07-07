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

//var key = []byte("secretKey")
/*
func randID() ([]byte, error) {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.New("creating new unique user ID error, " + err.Error())
	}
	return b, nil
}*/

func CreateUserID(cf config.Config) (string, string, error) {
	//id, err := randID()
	id := uuid.NewString()
	//idStr := hex.EncodeToString(id)
	//if err != nil {
	//	return "nil", "", err
	//}
	h := hmac.New(sha256.New, []byte(cf.AuthentKey) /*key*/)
	h.Write([]byte( /*idStr*/ id))
	signature := h.Sum(nil)
	//signatureStr := hex.EncodeToString(signature)
	var cookieSls []byte
	cookieSls = signature
	//for _, v := range []byte(id) {
	cookieSls = append(cookieSls, []byte(id)...)
	cookieStr := hex.EncodeToString(cookieSls)
	//}
	//encoder := new(base64.Encoding)
	//cookieStr := encoder.EncodeToString(cookieSls)
	fmt.Println("created id is " + id)
	fmt.Println(len([]byte(id)))
	fmt.Println("created signature is " + string(signature))
	fmt.Println(len(signature))
	fmt.Println("created sookieSls is " + string(cookieSls))
	fmt.Println(len([]byte(cookieSls)))
	fmt.Println("created cookieStr is " + cookieStr)
	fmt.Println(len([]byte(cookieStr)))
	return id /*signatureStr*/, cookieStr, nil
}

func CheckCookie(cookieStr string, cf config.Config) (string, error) {
	if cookieStr == "" {
		return "", errors.New("cookie is empty")
	}
	if !(len([]byte(cookieStr)) > 8) {
		return "", errors.New("wrong cookie len")
	}
	//encoder := new(base64.Encoding)
	gCookieSlc, err := hex.DecodeString(cookieStr)

	if err != nil {
		return "", fmt.Errorf("decoding gotted cookies error: %w", err)
	}
	fmt.Println("gotted decoded gCookieSlc is " + string(gCookieSlc))
	//decGotStr, err := hex.DecodeString(cookieStr)
	//if err != nil {
	//	return "", errors.New("decoding gotted cookie string err, " + err.Error())
	//}
	//gottedUserID := /*decGotStr*/ string([]byte(cookieStr)[64:])
	gottedUserID := /*decGotStr*/ string(gCookieSlc[32:])
	fmt.Println("gottedUserID is " + gottedUserID)
	//gottedUserIDStr := hex.EncodeToString(gottedUserID)
	//gottedSign := /*decGotStr*/ string([]byte(cookieStr)[:64])
	gottedSign := /*decGotStr*/ gCookieSlc[:32]
	fmt.Println("gottedSign is " + string(gottedSign))
	//gottedSignStr := hex.EncodeToString(gottedSign)
	h := hmac.New(sha256.New, []byte(cf.AuthentKey) /* key*/)
	h.Write([]byte( /*gottedUserIDStr*/ gottedUserID))
	signature := h.Sum(nil)
	//decSign := hex.EncodeToString(signature)

	//if hmac.Equal([]byte(decSign), []byte( /*gottedSignStr*/ gottedSign)) {
	if hmac.Equal(signature, gottedSign) {
		return /*gottedUserIDStr*/ gottedUserID /*gottedSignStr,*/, nil
	}
	return "", errors.New("wrong signature")
}
