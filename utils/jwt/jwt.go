package jsonWebToken

import (
	"go-api/config"
	"go-api/model"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	redis "go-api/utils/redis"
)

var timeExpiredTokenEmail = time.Minute * 15

type MapClaims struct {
	ID           string
	Email        string
	Role         model.Role
	Exp          float64
	IsSetupAdmin bool
}

var JWTKey = config.Env("JWT_KEY")

func CreateToken(args MapClaims) (string, error) {
	claims := jwt.MapClaims{
		"id":    args.ID,
		"email": args.Email,
		"role":  args.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	expiration, ok := claims["exp"].(float64)
	if !ok {
		return nil, err
	}

	if time.Now().Unix() > int64(expiration) {
		return nil, err
	}

	mapClaims := &MapClaims{
		ID:    claims["id"].(string),
		Email: claims["email"].(string),
		Role:  model.Role(claims["role"].(string)),
		Exp:   expiration,
	}

	return mapClaims, nil
}

var JWTKeyEmail = config.Env("JWT_KEY_EMAIL")

func CreateTokenEmail(args MapClaims) (string, error) {
	claims := jwt.MapClaims{
		"id":           args.ID,
		"email":        args.Email,
		"role":         args.Role,
		"exp":          time.Now().Add(timeExpiredTokenEmail).Unix(),
		"isSetupAdmin": args.IsSetupAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTKeyEmail))
	if err != nil {
		return "", err
	}

	AddToWhitelist(args.Email, tokenString)

	return tokenString, nil
}

func ParseTokenEmail(tokenString string) (*MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTKeyEmail), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	expiration, ok := claims["exp"].(float64)
	if !ok {
		return nil, err
	}

	if time.Now().Unix() > int64(expiration) {
		return nil, err
	}

	mapClaims := &MapClaims{
		ID:    claims["id"].(string),
		Email: claims["email"].(string),
		Role:  model.Role(claims["role"].(string)),
		Exp:   expiration,
	}

	return mapClaims, nil
}

func AddToWhitelist(email string, tokenString string) (bool, error) {
	result, err := redis.SetWithExpired(email, tokenString, timeExpiredTokenEmail)

	return result, err
}

func IsInWhitelist(email string, tokenString string) (bool, error) {
	result, err := redis.IsExisted(email, tokenString)

	return result, err
}

func RemoveFromWhitelist(email string) (bool, error) {
	result, err := redis.Delete(email)

	return result, err
}
