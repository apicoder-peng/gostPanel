package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Config JWT 配置
type Config struct {
	Secret string // 密钥
	Expire int64  // 过期时间（秒）
}

// JWT 实例
type JWT struct {
	config *Config
}

// 错误定义
var (
	ErrTokenExpired     = errors.New("token 已过期")
	ErrTokenNotValidYet = errors.New("token 尚未生效")
	ErrTokenMalformed   = errors.New("token 格式错误")
	ErrTokenInvalid     = errors.New("token 无效")
)

// New 创建 JWT 实例
func New(cfg *Config) *JWT {
	return &JWT{config: cfg}
}

// GenerateToken 生成 Token
func (j *JWT) GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()
	expireAt := now.Add(time.Duration(j.config.Expire) * time.Second)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gost-panel",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ParseToken 解析 Token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		// 判断具体错误类型
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新 Token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新 Token
	return j.GenerateToken(claims.UserID, claims.Username)
}
