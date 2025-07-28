// common/crypto.go
package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func GenerateJWTToken(payload map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{}

	for k, v := range payload {
		claims[k] = v
	}

	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(2 * time.Hour).Unix()
	}

	if _, ok := claims["iat"]; !ok {
		claims["iat"] = time.Now().Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(CryptoSecret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func GenerateHMACWithKey(key []byte, data string) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateHMAC(data string) string {
	h := hmac.New(sha256.New, []byte(CryptoSecret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateDeviceIDWithIP(ip string) string {
	// 1) 隨機 8 byte
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		SysError("Failed to generate random bytes for device ID: " + err.Error())
		return "" // 失敗就回空
	}

	// 2) IP 的 SHA-256 前 8 byte
	h := sha256.Sum256([]byte(ip))
	ipPart := h[:8]

	// 3) 拼接並回傳
	id := append(randBytes, ipPart...)
	return hex.EncodeToString(id)
}

func ValidateUser(token string, loginIp string) bool {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return []byte(CryptoSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		SysDebug("Invalid token: " + err.Error())
		return false
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if ip, ok := claims["Login_IP"].(string); ok {
			if ip != loginIp {
				return false // IP 不匹配
			}
			return true // IP 匹配
		}
	} else {
		SysDebug("claim conversion error")
		return false // 轉換 claims 失敗
	}

	return true
}

func GetJWTPayload(token string) (map[string]interface{}, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return []byte(CryptoSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, err // 解析失敗或無效
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		payload := make(map[string]interface{})
		for k, v := range claims {
			payload[k] = v
		}
		return payload, nil
	}

	return nil, jwt.NewValidationError("invalid token claims", jwt.ValidationErrorClaimsInvalid)
}
