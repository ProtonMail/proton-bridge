package bridge

// IsAuthorized returns whether the user has received an Auth from the API yet.
func (u *User) IsAuthorized() bool {
	return u.isAuthorized
}
