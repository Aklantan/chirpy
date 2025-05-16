package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
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

func TestJWT(t *testing.T){
	secret := "walrider"
	testID := uuid.New()
	token, err := MakeJWT(testID,secret,30*time.Minute)
	if err != nil {
		t.Errorf("cannot make JWT : %v ",err)
	}
	_, err = ValidateJWT(token,secret)
	if err != nil{
		t.Errorf("UUIDs do not match : %v", err)
	}
}