package security

import (
	"crypto/rand"

	"github.com/gofrs/uuid"
)

// GenerateRandomUUIDV5 will return a 32bit random seeded UUID based on
// a randomly generated UUID v4.
func GenerateRandomUUIDV5() string {
	nsUUID, _ := uuid.NewV4()
	token := make([]byte, 32)
	_, _ = rand.Read(token)
	namespace := string(token)
	t := uuid.NewV5(nsUUID, namespace)
	return t.String()
}
