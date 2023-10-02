package totp

import (
	"botp-gateway/config"
	"encoding/base32"
	"log"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var TOTP_PREFIX_KEY = config.Env("TOTP_PREFIX_KEY")

var ValidateOpts = totp.ValidateOpts{
	Period:    900, // 15 minutes
	Skew:      0,
	Digits:    4,
	Algorithm: otp.AlgorithmSHA1,
}

func GenerateCode(email string) (string, error) {

	key := TOTP_PREFIX_KEY + email

	secretBytes := []byte(key)

	// Encode the byte slice as base32
	encodedSecret := base32.StdEncoding.EncodeToString(secretBytes)

	return totp.GenerateCodeCustom(encodedSecret, time.Now().UTC(), ValidateOpts)
}

func VerifyCode(email string, passcode string) (bool, error) {

	key := TOTP_PREFIX_KEY + email

	secretBytes := []byte(key)

	// Encode the byte slice as base32
	encodedSecret := base32.StdEncoding.EncodeToString(secretBytes)

	isValid, err := totp.ValidateCustom(passcode, encodedSecret, time.Now().UTC(), ValidateOpts)
	if err != nil {
		log.Println(err)
	}

	return isValid, err
}
