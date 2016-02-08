package utils

import (
	"github.com/satori/go.uuid"
	"strings"
)

func GenerateUUID() string {
	sig := uuid.NewV1().String()
	sig = strings.Replace(sig, "-", "", -1)
	return sig
}
