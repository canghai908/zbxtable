package auth

import (
	"errors"
	"fmt"

	"zbxtable/internal/model"

	"github.com/golang-jwt/jwt/v4"
)

// 与 model.GetEncryptionKey 回退值保持一致：安装后应使用数据库中的随机密钥，
// 禁止继续用公开默认值签发 / 校验令牌。
const defaultEncryptionKey = "zbxtable-default-encryption-key-2024"

// IssueToken 使用安装级密钥签发 HS256 JWT（iss = 用户名）。
func IssueToken(username string, expiresUnix int64) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}
	claims := &jwt.StandardClaims{
		ExpiresAt: expiresUnix,
		Issuer:    username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken 校验 JWT：仅接受 HMAC 族算法，拒绝 none / RSA 等。
// 成功时返回 issuer（用户名）。
func ValidateToken(tokenString string) (bool, string, error) {
	if tokenString == "" {
		return false, "", errors.New("token is empty")
	}
	secret, err := jwtSecret()
	if err != nil {
		return false, "", err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return false, "", err
	}
	if token == nil || !token.Valid {
		return false, "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", errors.New("invalid token claims")
	}
	iss, _ := claims["iss"].(string)
	if iss == "" {
		return false, "", errors.New("missing issuer")
	}
	return true, iss, nil
}

func jwtSecret() ([]byte, error) {
	key := model.GetEncryptionKey()
	if key == "" || key == defaultEncryptionKey {
		return nil, errors.New("jwt secret not configured; complete installation first")
	}
	return []byte(key), nil
}
