package util

import "math/rand"

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(seed int64, length int) string {
	rand.Seed(seed)
	str := ""
	for i := 0; i < length; i++ {
		str += string(chars[rand.Int31n(int32(len(chars)))])
	}

	return str
}
