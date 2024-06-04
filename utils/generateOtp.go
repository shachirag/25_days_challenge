package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func Generate6DigitOtp() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	randInt := r.Intn(900000) + 100000
	return strconv.Itoa(randInt)
}
