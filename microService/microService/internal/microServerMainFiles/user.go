package microServerMainFiles

// User defines the structure for an API user
type User struct {
	Email    string `bson:"email"`    // Email of the user, used as a unique identifier
	Password string `bson:"password"` // Password of the user, which should be securely hashed
}
