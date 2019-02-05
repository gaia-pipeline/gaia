package security

import (
	"crypto/rand"

	uuid "github.com/satori/go.uuid"
)

// GenerateRandomUUIDV5 will return a 32bit random seeded UUID based on
// a randomly generated UUID v4.
func GenerateRandomUUIDV5() string {
	nsUUID := uuid.NewV4()
	token := make([]byte, 32)
	rand.Read(token)
	namespace := string(token)
	t := uuid.NewV5(nsUUID, namespace)
	return t.String()
}
