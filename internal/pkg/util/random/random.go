package random

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Int generates a random number between min and max.
func Int(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// From0To1000 generates a random number between 0 and 1000.
func From0To1000() int64 {
	return Int(0, 1000)
}

// From1To1000 generates a random number between 1 and 1000.
func From1To1000() int64 {
	return Int(1, 1000)
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// String generates a random string of length n.
func String(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}
