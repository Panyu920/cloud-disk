package utils

import "regexp"

var (
	PhoneRegex = regexp.MustCompile(`^1[3456789]\d{9}$`)
)

func IsValidPhone(phone string) bool {
	return PhoneRegex.MatchString(phone)
}
