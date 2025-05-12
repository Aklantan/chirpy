package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	passw := "hangGlider"
	hash, err := HashPassword(passw)
	if err != nil {
		t.Errorf("Password %s could not be hashed", passw)
	}
	err = CheckPasswordHash(hash, passw)
	if err != nil {
		t.Error("Passwords did not match")
	}

}
