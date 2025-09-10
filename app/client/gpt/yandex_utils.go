package gpt

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

func (g *YandexGpt) signedJWTToken() (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    g.cfg.Yandex.ServiceAccountID,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	token.Header["kid"] = g.cfg.Yandex.KeyID

	privateKey, err := g.loadPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

func (g *YandexGpt) loadPrivateKey() (*rsa.PrivateKey, error) {
	keyData, err := g.readPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return rsaPrivateKey, nil
}

func (g *YandexGpt) readPrivateKey() (*iamkey.Key, error) {
	var keyData *iamkey.Key
	if err := json.Unmarshal([]byte(g.cfg.Yandex.Key), &keyData); err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return keyData, nil
}

func (g *YandexGpt) getIAMToken(ctx context.Context) (string, error) {
	authKey, err := g.readPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}

	credentials, err := ycsdk.ServiceAccountKey(authKey)
	if err != nil {
		return "", fmt.Errorf("could not get service account key: %w", err)
	}

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: credentials,
	})
	if err != nil {
		return "", fmt.Errorf("could not build sdk: %w", err)
	}

	jwtToken, err := g.signedJWTToken()
	if err != nil {
		return "", fmt.Errorf("could not get token: %w", err)
	}

	iamRequest := &iam.CreateIamTokenRequest{
		Identity: &iam.CreateIamTokenRequest_Jwt{Jwt: jwtToken},
	}

	newKey, err := sdk.IAM().IamToken().Create(ctx, iamRequest)
	if err != nil {
		return "", fmt.Errorf("could not create IAM token: %w", err)
	}

	return newKey.IamToken, nil
}
