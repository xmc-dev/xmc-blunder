package util

// DetailedScope contains information about a scope
type DetailedScope struct {
	Title    string
	Synopsis string
}

// ConsentRequestData contains information that is displayed to the user
// when requested to accept access by an app
type ConsentRequestData struct {
	ClientName          string
	OwnerClientID       string
	SubjectClientID     string
	Scopes              []DetailedScope
	ResponseType        string
	ClientID            string
	RedirectURI         string
	State               string
	Scope               string
	Prefix              string
	CodeChallenge       string
	CodeChallengeMethod string
}
