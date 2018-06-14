package users

import (
	"log"
	"time"

	"github.com/almerlucke/go-utils/server/auth/password"
	"github.com/almerlucke/go-utils/sql/database"
	"github.com/almerlucke/go-utils/sql/model"
	"github.com/almerlucke/go-utils/sql/types"

	"github.com/satori/go.uuid"
)

// TokenRequestErrorCode response code for token request
type TokenRequestErrorCode int

// LoginErrorCode response code for login
type LoginErrorCode int

const (
	// MaxLoginAttempts maximum number of wrong login attempts
	MaxLoginAttempts = 3

	// RequestExpiryHours number of hours before expiry
	RequestExpiryHours = 10
)

const (
	// TokenRequestErrorCodeUnknown unknown
	TokenRequestErrorCodeUnknown TokenRequestErrorCode = iota
	// TokenRequestErrorCodeSuccess success
	TokenRequestErrorCodeSuccess
	// TokenRequestErrorCodeExpired expired
	TokenRequestErrorCodeExpired
	// TokenRequestErrorCodeUnknownToken unknown token
	TokenRequestErrorCodeUnknownToken
	// TokenRequestErrorCodeUnknownUser unknown user
	TokenRequestErrorCodeUnknownUser
)

const (
	// LoginErrorCodeUnknown unknown
	LoginErrorCodeUnknown LoginErrorCode = iota
	// LoginErrorCodeSuccess success
	LoginErrorCodeSuccess
	// LoginErrorCodeUnknownUser unknown user
	LoginErrorCodeUnknownUser
	// LoginErrorCodeWrongPassword wrong password
	LoginErrorCodeWrongPassword
	// LoginErrorCodeBlocked blocked for too many attempts
	LoginErrorCodeBlocked
)

const (
	// PasswordResetRequestType for password reset request
	PasswordResetRequestType = "password_reset"
)

// MinimumProfile model to be embedded
type MinimumProfile struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Avatar      string `json:"avatar" db:"avatar"`
}

// User model
type User struct {
	model.Model
	MinimumProfile
	Username           string `json:"username" db:"username"`
	Email              string `json:"email" db:"email"`
	Password           string `json:"-" db:"password"`
	LoginAttempts      int    `json:"-" db:"login_attempts" sql:"default 0"`
	EmailConfirmed     bool   `json:"-" db:"email_confirmed" sql:"default 0"`
	EnabledTwoFactor   bool   `json:"-" db:"enabled_two_factor" sql:"default 0"`
	ValidatedTwoFactor bool   `json:"-" db:"validated_two_factor" sql:"default 0"`
	TOTP               []byte `json:"-" db:"totp"`
}

// BelongsTo model to store mapping between user and organization
type BelongsTo struct {
	model.Model
	UserID         int64        `db:"user_id"`
	OrganizationID int64        `db:"organization_id"`
	Role           types.String `db:"role" sql:"override,varchar(32)"`
}

// Organization model
type Organization struct {
	model.Model
	MinimumProfile
}

// Request model
type Request struct {
	model.Model
	ExpiryDate       types.DateTime `json:"expiryDate" db:"expiry_date"`
	Token            string         `json:"token" db:"token" sql:"override,varchar(128)"`
	OrganizationID   int64          `json:"-" db:"organization_id"`
	OrganizationName string         `json:"organizationName" db:"organization_name"`
	InvitedBy        string         `json:"invitedBy" db:"invited_by"`
	InvitedByID      int64          `json:"-" db:"invited_by_id"`
	Email            string         `json:"email" db:"email" sql:"override,varchar(32)"`
	Username         string         `json:"username" db:"username"`
	ExistingUserID   int64          `json:"-" db:"existing_user_id"`
	Role             string         `json:"role" db:"role" sql:"override,varchar(32)"`
	Type             string         `json:"type" db:"type" sql:"override,varchar(32)"`
}

// UserTable user table
var UserTable model.Tabler

// BelongsToTable user to organization mapping table
var BelongsToTable model.Tabler

// OrganizationTable organization table
var OrganizationTable model.Tabler

// RequestTable stores invitation, password change and ownership change requests
var RequestTable model.Tabler

// Initialize tables
func init() {
	table, err := model.NewTable("users", &User{})
	if err != nil {
		log.Fatalf("error creating user table: %v", err)
	}

	UserTable = table

	table, err = model.NewTable("user_organization_mapping", &BelongsTo{})
	if err != nil {
		log.Fatalf("error creating user organization mapping table: %v", err)
	}

	BelongsToTable = table

	table, err = model.NewTable("organizations", &Organization{})
	if err != nil {
		log.Fatalf("error creating organization table: %v", err)
	}

	OrganizationTable = table

	table, err = model.NewTable("requests", &Request{})
	if err != nil {
		log.Fatalf("error creating request table: %v", err)
	}

	RequestTable = table
}

// LoginWithEmailOrUsername find a user by username or email and verify password.
// Returns if we found a user, if the password was correct and if an error occurred
func LoginWithEmailOrUsername(identity string, pwd string, queryer database.Queryer) (*User, LoginErrorCode, error) {
	result, err := UserTable.Select("*").Where("{{Username}}=? OR {{Email}}=?").Run(queryer, identity, identity)
	if err != nil {
		return nil, LoginErrorCodeUnknown, err
	}

	users := result.([]*User)
	if len(users) == 0 {
		return nil, LoginErrorCodeUnknownUser, nil
	}

	user := users[0]

	if user.LoginAttempts >= MaxLoginAttempts {
		return user, LoginErrorCodeBlocked, nil
	}

	if password.CheckHashAndPassword(user.Password, pwd) {
		// Correct login, reset login attempts
		user.LoginAttempts = 0

		_, err = UserTable.Update(user, queryer)
		if err != nil {
			return nil, LoginErrorCodeUnknown, err
		}

		return user, LoginErrorCodeSuccess, nil
	}

	// Wrong password, increment login attempts
	user.LoginAttempts = user.LoginAttempts + 1

	_, err = UserTable.Update(user, queryer)
	if err != nil {
		return nil, LoginErrorCodeUnknown, err
	}

	return user, LoginErrorCodeWrongPassword, nil
}

// IsUsernameAvailable checks if a username is available
func IsUsernameAvailable(username string, queryer database.Queryer) (bool, error) {
	result, err := UserTable.Select("{{ID}}").Where("{{Username}}=?").Run(queryer, username)
	if err != nil {
		return false, err
	}

	return len(result.([]*User)) == 0, nil
}

// IsEmailAvailable checks if an email is available
func IsEmailAvailable(email string, queryer database.Queryer) (bool, error) {
	result, err := UserTable.Select("{{ID}}").Where("{{Email}}=?").Run(queryer, email)
	if err != nil {
		return false, err
	}

	return len(result.([]*User)) == 0, nil
}

// RegisterUser register a user
func RegisterUser(user *User, queryer database.Queryer) error {
	_, err := UserTable.Insert([]interface{}{user}, queryer)
	return err
}

// GenerateExpiryDate generate an expiry date hours from now
func GenerateExpiryDate(hours int) time.Time {
	return time.Now().UTC().Add(time.Duration(hours) * time.Second)
}

// GeneratePasswordResetRequest generate and insert a password reset request
func GeneratePasswordResetRequest(userID int64, queryer database.Queryer) (*Request, error) {
	request := &Request{
		Token:          uuid.NewV4().String(),
		Type:           PasswordResetRequestType,
		ExistingUserID: userID,
		ExpiryDate:     types.DateTime(GenerateExpiryDate(RequestExpiryHours)),
	}

	_, err := RequestTable.Insert([]interface{}{request}, queryer)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// ValidatePasswordResetRequest validate a new password request with token
func ValidatePasswordResetRequest(token string, newPassword string, queryer database.Queryer) (TokenRequestErrorCode, error) {
	result, err := RequestTable.Select("*").Where("{{Token}}=?").Run(queryer, token)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	requests := result.([]*Request)
	if len(requests) == 0 {
		return TokenRequestErrorCodeUnknownToken, nil
	}

	request := requests[0]

	if time.Now().UTC().After(time.Time(request.ExpiryDate)) {
		return TokenRequestErrorCodeExpired, nil
	}

	result, err = UserTable.Select("*").Where("{{ID}}=?").Run(queryer, request.ExistingUserID)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	users := result.([]*User)
	if len(users) == 0 {
		return TokenRequestErrorCodeUnknownUser, err
	}

	user := users[0]
	pwd, err := password.GetPasswordHash(newPassword)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	user.Password = pwd
	user.LoginAttempts = 0

	_, err = UserTable.Update(user, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	_, err = RequestTable.Delete(request, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	return TokenRequestErrorCodeSuccess, nil
}

/*
log in
sign up
create organization -> become admin
invite people
manage organization (name? profile not in scope)
manage people (invite, cancel invite, delete)
exchange ownership
forgot password
reset password
*/
