package security

import "testing"

func TestRandomGeneratingUUIDV5(t *testing.T) {
	uuid1 := GenerateRandomUUIDV5()
	uuid2 := GenerateRandomUUIDV5()
	if uuid1 == uuid2 {
		t.Fatalf("the two random generated uuids should not have equalled: uuid1: %s, uuid2: %s\n", uuid1, uuid2)
	}
}
