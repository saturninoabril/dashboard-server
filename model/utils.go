package model

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/base32"
	mathRand "math/rand"
	"strconv"

	"github.com/pborman/uuid"
)

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

// NewID is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off.
func NewID() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	_, _ = encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

// NewRandomString returns a random string of the given length.
// The resulting entropy will be (5 * length) bits.
func NewRandomString(length int) string {
	data := make([]byte, 1+(length*5/8))
	_, _ = cryptoRand.Read(data)
	return encoding.EncodeToString(data)[:length]
}

// NewRandomNumber returns a random string of the given length, built from digits
func NewRandomNumber(length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s += strconv.Itoa(mathRand.Intn(10))
	}
	return s
}
