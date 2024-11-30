package jwt

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Service struct {
	secret []byte
}

func New(secret string) *Service {
	return &Service{
		secret: []byte(secret),
	}
}

func (s *Service) GenerateTokenPair(guid string, ip string, duration time.Duration) (*Tokens, error) {
	var err error
	tokens := &Tokens{}

	accessToken := jwt.New(jwt.SigningMethodHS512)

	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["guid"] = guid
	accessClaims["ip"] = ip
	accessClaims["type"] = "access"
	accessClaims["exp"] = time.Now().Add(duration).Unix()

	tokens.AccessToken, err = accessToken.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	// type_accessToken_ip
	refreshToken := "refresh_" + tokens.AccessToken + "_" + ip

	hash := sha256.Sum256([]byte(refreshToken))

	tokens.RefreshToken = base64.URLEncoding.EncodeToString(hash[:])

	return tokens, nil
}

func (s *Service) ValidateAccessToken(accessToken string) (string, string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.secret, nil
	})
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "access" {
		return "", "", errors.New("invalid token")
	}

	return claims["ip"].(string), claims["guid"].(string), nil
}

func (s *Service) ValidateRefreshToken(refreshToken string, accessToken string, ip string) error {
	expectedRefreshToken := "refresh_" + accessToken + "_" + ip

	hash := sha256.Sum256([]byte(expectedRefreshToken))

	finalRefreshToken := base64.URLEncoding.EncodeToString(hash[:])

	if refreshToken != finalRefreshToken {
		return errors.New("invalid refresh token")
	}

	return nil
}
