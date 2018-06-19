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

const (
	// OwnerRole organization owner
	OwnerRole = "owner"

	// AdminRole organization admin
	AdminRole = "admin"
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

	// ConfirmEmailExpiryHours number of hours before email confirmation expiry
	ConfirmEmailExpiryHours = 48

	// InvitationExpiryHours number of hours before invitation expiry
	InvitationExpiryHours = 48
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
	// LoginErrorEmailUnconfirmed email is not yet confirmed
	LoginErrorEmailUnconfirmed
)

const (
	// PasswordResetRequestType password reset request
	PasswordResetRequestType = "password_reset"

	// ConfirmEmailRequestType confirm email request
	ConfirmEmailRequestType = "confirm_email"

	// InvitationRequestType invitation request
	InvitationRequestType = "invitation"
)

// MinimumProfile model to be embedded
type MinimumProfile struct {
	Name        string       `json:"name" db:"name"`
	Description types.String `json:"description" db:"description"`
	Avatar      types.String `json:"avatar" db:"avatar"`
}

// User model
type User struct {
	model.Model
	MinimumProfile
	Username           string `json:"username" db:"username" sql:"override,varchar(64)"`
	Email              string `json:"email" db:"email" sql:"override,varchar(256)"`
	Password           string `json:"-" db:"password"`
	LoginAttempts      int    `json:"-" db:"login_attempts" sql:"override,tinyint default 0"`
	EmailConfirmed     bool   `json:"-" db:"email_confirmed" sql:"default 0"`
	EnabledTwoFactor   bool   `json:"-" db:"enabled_two_factor" sql:"default 0"`
	ValidatedTwoFactor bool   `json:"-" db:"validated_two_factor" sql:"default 0"`
	TOTP               []byte `json:"-" db:"totp"`
}

// BelongsTo model to store mapping between user and organization
type BelongsTo struct {
	model.Model
	UserID         uint64       `db:"user_id"`
	OrganizationID uint64       `db:"organization_id"`
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
	OrganizationID   uint64         `json:"-" db:"organization_id"`
	OrganizationName string         `json:"organizationName" db:"organization_name"`
	InvitedBy        string         `json:"invitedBy" db:"invited_by"`
	InvitedByID      uint64         `json:"-" db:"invited_by_id"`
	Email            string         `json:"email" db:"email" sql:"override,varchar(256)"`
	Username         string         `json:"username" db:"username" sql:"override,varchar(64)"`
	ExistingUserID   uint64         `json:"-" db:"existing_user_id"`
	Role             types.String   `json:"role" db:"role" sql:"override,varchar(32)"`
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

	table.KeysAndConstraints = []string{
		"KEY `username` (`username`)",
		"KEY `email` (`email`)",
	}

	UserTable = table

	table, err = model.NewTable("user_organization_mapping", &BelongsTo{})
	if err != nil {
		log.Fatalf("error creating user organization mapping table: %v", err)
	}

	table.KeysAndConstraints = []string{
		"KEY `user_id` (`user_id`)",
		"KEY `organization_id` (`organization_id`)",
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

	table.KeysAndConstraints = []string{
		"KEY `token` (`token`)",
	}

	RequestTable = table
}

// BelongsToOrganization check if user belongs to organization
func (user *User) BelongsToOrganization(organizationID uint64, queryer database.Queryer) (*BelongsTo, error) {
	result, err := BelongsToTable.Select("{{ID}}").Where("{{OrganizationID}}=? AND {{UserID}}=?").Run(queryer, organizationID, user.ID)
	if err != nil {
		return nil, err
	}

	connections := result.([]*BelongsTo)
	if len(connections) == 0 {
		return nil, nil
	}

	return connections[0], nil
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

	if !user.EmailConfirmed {
		return user, LoginErrorEmailUnconfirmed, nil
	}

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
func RegisterUser(user *User, queryer database.Queryer) (*Request, error) {
	result, err := UserTable.Insert([]interface{}{user}, queryer)
	if err != nil {
		return nil, err
	}

	userID, _ := result.LastInsertId()

	request := &Request{
		Token:          uuid.NewV4().String(),
		Type:           ConfirmEmailRequestType,
		ExistingUserID: uint64(userID),
		ExpiryDate:     types.DateTime(GenerateExpiryDate(ConfirmEmailExpiryHours)),
	}

	// Create confirm email request
	_, err = RequestTable.Insert([]interface{}{request}, queryer)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// GetRequestForToken get request for token
func GetRequestForToken(token string, queryer database.Queryer) (*Request, error) {
	result, err := RequestTable.Select("*").Where("{{Token}}=?").Run(queryer, token)
	if err != nil {
		return nil, err
	}

	requests := result.([]*Request)
	if len(requests) == 0 {
		return nil, nil
	}

	return requests[0], nil
}

// ValidateExistingUserTokenRequest validate token requests for existing users
func ValidateExistingUserTokenRequest(token string, deleteRequest bool, queryer database.Queryer) (TokenRequestErrorCode, *User, error) {
	request, err := GetRequestForToken(token, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, nil, err
	}

	if request == nil {
		return TokenRequestErrorCodeUnknownToken, nil, nil
	}

	// Select user for request
	result, err := UserTable.Select("*").Where("{{ID}}=?").Run(queryer, request.ExistingUserID)
	if err != nil {
		return TokenRequestErrorCodeUnknown, nil, err
	}

	users := result.([]*User)
	if len(users) == 0 {
		return TokenRequestErrorCodeUnknownUser, nil, err
	}

	user := users[0]

	// Check if the request is expired
	if time.Now().UTC().After(time.Time(request.ExpiryDate)) {
		return TokenRequestErrorCodeExpired, user, nil
	}

	if deleteRequest {
		_, err = RequestTable.Delete(request, queryer)
		if err != nil {
			return TokenRequestErrorCodeUnknown, nil, err
		}
	}

	return TokenRequestErrorCodeSuccess, user, nil
}

// ConfirmRegistration confirm user registration email
func ConfirmRegistration(token string, queryer database.Queryer) (TokenRequestErrorCode, *User, error) {
	code, user, err := ValidateExistingUserTokenRequest(token, true, queryer)
	if err != nil || code != TokenRequestErrorCodeSuccess {
		return code, user, err
	}

	user.EmailConfirmed = true

	_, err = UserTable.Update(user, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, nil, err
	}

	return TokenRequestErrorCodeSuccess, user, nil
}

// GenerateExpiryDate generate an expiry date hours from now
func GenerateExpiryDate(hours int) time.Time {
	return time.Now().UTC().Add(time.Duration(hours) * time.Second)
}

// GeneratePasswordResetRequest generate and insert a password reset request
func GeneratePasswordResetRequest(userID uint64, queryer database.Queryer) (*Request, error) {
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
func ValidatePasswordResetRequest(token string, newPassword string, queryer database.Queryer) (TokenRequestErrorCode, *User, error) {
	code, user, err := ValidateExistingUserTokenRequest(token, true, queryer)
	if err != nil || code != TokenRequestErrorCodeSuccess {
		return code, user, err
	}

	// Get new password hash
	pwd, err := password.GetPasswordHash(newPassword)
	if err != nil {
		return TokenRequestErrorCodeUnknown, nil, err
	}

	user.Password = pwd
	user.LoginAttempts = 0

	// Update user
	_, err = UserTable.Update(user, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, nil, err
	}

	return TokenRequestErrorCodeSuccess, user, nil
}

// GetUserWithUsername get a user by username
func GetUserWithUsername(username string, queryer database.Queryer) (*User, error) {
	result, err := UserTable.Select("{{ID}}").Where("{{Username}}=?").Run(queryer, username)
	if err != nil {
		return nil, err
	}

	users := result.([]*User)

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// InviteExistingUserToOrganization creates an invitation request and adds it to the db
func InviteExistingUserToOrganization(user *User, invitedBy *User, organization *Organization, role string, queryer database.Queryer) (*Request, error) {
	request := &Request{
		Token:            uuid.NewV4().String(),
		Type:             InvitationRequestType,
		ExpiryDate:       types.DateTime(GenerateExpiryDate(InvitationExpiryHours)),
		OrganizationID:   organization.ID,
		OrganizationName: organization.Name,
		InvitedBy:        invitedBy.Name,
		InvitedByID:      invitedBy.ID,
		Username:         user.Username,
		Role:             types.String(role),
		ExistingUserID:   user.ID,
	}

	_, err := RequestTable.Insert([]interface{}{request}, queryer)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// InviteNewUserToOrganization creates an invitation request and adds it to the db
func InviteNewUserToOrganization(emailAddress string, invitedBy *User, organization *Organization, role string, queryer database.Queryer) (*Request, error) {
	request := &Request{
		Token:            uuid.NewV4().String(),
		Type:             InvitationRequestType,
		ExpiryDate:       types.DateTime(GenerateExpiryDate(InvitationExpiryHours)),
		OrganizationID:   organization.ID,
		OrganizationName: organization.Name,
		InvitedBy:        invitedBy.Name,
		InvitedByID:      invitedBy.ID,
		Username:         "",
		Role:             types.String(role),
		ExistingUserID:   0,
		Email:            emailAddress,
	}

	_, err := RequestTable.Insert([]interface{}{request}, queryer)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// AcceptInvitation accept an invitation and add user to the organization
func AcceptInvitation(token string, user *User, queryer database.Queryer) (TokenRequestErrorCode, error) {
	request, err := GetRequestForToken(token, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	if request == nil {
		return TokenRequestErrorCodeUnknownToken, nil
	}

	// Check if the request is expired
	if time.Now().UTC().After(time.Time(request.ExpiryDate)) {
		return TokenRequestErrorCodeExpired, nil
	}

	// Make a connection
	belongsTo := &BelongsTo{
		Role:           request.Role,
		OrganizationID: request.OrganizationID,
		UserID:         user.ID,
	}

	_, err = BelongsToTable.Insert([]interface{}{belongsTo}, queryer)
	if err != nil {
		return TokenRequestErrorCodeUnknown, err
	}

	// Delete request
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
