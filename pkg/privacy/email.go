package privacy

import "strings"

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid_email"
	}
	username := parts[0]
	if len(username) < 2 {
		return username + "****@" + parts[1]
	}
	return username[:2] + "****@" + parts[1]
}
