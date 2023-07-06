package authentication

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

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
	signatureStr := hex.EncodeToString(signature)
	return /*idStr*/ id, signatureStr /*string(signature)*/, nil
}

func CheckCookie(cookieStr string, cf config.Config) (string, error) {
	if cookieStr == "" {
		return "", errors.New("cookie is empty")
	}
	if !(len([]byte(cookieStr)) > 8) {
		return "", errors.New("wrong cookie len")
	}
	//decGotStr, err := hex.DecodeString(cookieStr)
	//if err != nil {
	//	return "", errors.New("decoding gotted cookie string err, " + err.Error())
	//}
	gottedUserID := /*decGotStr*/ cookieStr[32:]
	//gottedUserIDStr := hex.EncodeToString(gottedUserID)
	gottedSign := /*decGotStr*/ cookieStr[:32]
	//gottedSignStr := hex.EncodeToString(gottedSign)
	h := hmac.New(sha256.New, []byte(cf.AuthentKey) /* key*/)
	h.Write([]byte( /*gottedUserIDStr*/ gottedUserID))
	signature := h.Sum(nil)
	decSign := hex.EncodeToString(signature)

	if hmac.Equal([]byte(decSign), []byte( /*gottedSignStr*/ gottedSign)) {
		return /*gottedUserIDStr*/ gottedUserID /*gottedSignStr,*/, nil
	}
	return "", errors.New("wrong signature")
}
