package privacy

import "testing"

func TestPassword(t *testing.T) {
	passwords := []string{"", "12", "password_1234"}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			salt := "salt"
			hashPassword := HashPassword(password, salt)
			if !VerifyPassword(password, hashPassword, salt) {
				t.Error("password dont match")
			}
		})
	}
}
