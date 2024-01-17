package vulinbox

import (
	"github.com/jinzhu/gorm"
	"math/rand"
)

type VulinUser struct {
	gorm.Model

	Username string
	Password string
	Age      int

	Role string // Add a role field

	Remake string // adds a note field

}

// Generate a specified amount of random user data
func generateRandomUsers(count int) []VulinUser {
	// Define optional username and password characters
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate test data
	users := make([]VulinUser, count)
	for i := 0; i < count; i++ {
		// Generate a random username and password
		username := generateRandomString(chars, 8)
		password := generateRandomString(chars, 12)

		// generates a random age (between 18-65 years old)
		age := rand.Intn(48) + 18

		// Create a user instance and assign it Add to the user list
		users[i] = VulinUser{
			Username: username,
			Password: password,
			Age:      age,
			Role:     "user",
			Remake:   "I am the user",
		}
	}

	return users
}

// Generate a random string of specified length
func generateRandomString(chars string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
