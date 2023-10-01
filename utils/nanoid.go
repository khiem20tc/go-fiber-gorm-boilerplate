package utils

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func GenerateNanoID() (string, error) {
	return gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)
}

func GenerateReferralCode() string {
	part1, _ := gonanoid.Generate("0123456789", 2)
	part2, _ := gonanoid.Generate("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 2)
	part3, _ := gonanoid.Generate("0123456789", 2)

	referralCode := part1 + part2 + part3

	return referralCode
}
