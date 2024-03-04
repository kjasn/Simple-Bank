package utils

import (
	"math/rand"
	"strings"
	"time"
)

var alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"


func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomString generate a random string of length n
func RandomString(n int) string {
	var result strings.Builder	// using strings.Builder rather than string
	sz := len(alphabet)
	for i := 0; i < n; i++ {
		ch := alphabet[rand.Intn(sz)]
		result.WriteByte(ch)
	}
	return result.String()
}


// RandomInt generate a random int64   [mn, mx]
func RandomInt(mn, mx int64) int64 {	// in case have confilcs with func max & min
	return mn + rand.Int63n(mx - mn + 1)
}


// RandomOwner generate a random owner name with 6 character
func RandomOwner() string {
	return RandomString(6)	// 
}


// RandomMoney generate a random money amount
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generate a random currency code
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "RMB"}	
	n := len(currencies)
	return currencies[rand.Intn(n)]
}