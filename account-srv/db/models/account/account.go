package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
)

// Type is the user type.
type Type int32

const (
	// USER means that the Account belongs to a human
	USER Type = 0

	// SERVICE means that the Account belongs to a machine
	SERVICE Type = 1
)

// Account represents a user/service account in the database
type Account struct {
	UUID   uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	RoleID string    `gorm:"default:null"`

	Type Type

	ClientID     string
	ClientSecret string
	Name         string

	OwnerUUID    uuid.UUID `gorm:"type:uuid"`
	CallbackURL  string
	IsFirstParty bool
	IsPublic     bool
	Scope        string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ToProto represents the model as a protobuf message
func (acc *Account) ToProto() *account.Account {
	pa := &account.Account{
		Uuid:   acc.UUID.String(),
		RoleId: acc.RoleID,

		Type: account.Type(acc.Type),

		ClientId:     acc.ClientID,
		ClientSecret: acc.ClientSecret,
		Name:         acc.Name,

		OwnerUuid:    acc.OwnerUUID.String(),
		CallbackUrl:  acc.CallbackURL,
		IsFirstParty: acc.IsFirstParty,
		IsPublic:     acc.IsPublic,
		Scope:        acc.Scope,

		CreatedAt: acc.CreatedAt.Unix(),
		UpdatedAt: acc.UpdatedAt.Unix(),
	}
	if acc.Type == USER {
		pa.OwnerUuid = ""
	}

	return pa
}
