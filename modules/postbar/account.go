package postbar

import (
	"github.com/purstal/pbtools/modules/misc"
)

type Account struct {
	ID string

	BDUSS string

	NetType       string
	ClientType    string
	ClientID      string
	ClientVersion string
	PhoneIMEI     string
}

const (
	Windows8 = `4`
	Android  = `2`
)

var DefaultPhoneIMEI = misc.ComputeMD5(``)

func NewDefaultAndroidAccount(id string) *Account {
	return &Account{
		ID:            id,
		NetType:       ``,
		ClientType:    Android,
		ClientID:      ``,
		ClientVersion: `6.1.3`,
		PhoneIMEI:     DefaultPhoneIMEI, //...
	}

}

//Maribel Hearn ---------v
const DefaultClientID = `4C-07-16-00-F1-C0-5B-47-62-86-B7-35-AF-24-24-DB-E7-05-86-8B-BF-E6-A4-06-B2-54-E3-AB-81-2D-9D-32`

func NewDefaultWindows8Account(id string) *Account {
	return &Account{
		ID:            id,
		NetType:       `3`,
		ClientType:    Windows8,
		ClientID:      DefaultClientID,
		ClientVersion: `1.5.0.0`,
		PhoneIMEI:     DefaultPhoneIMEI,
	}
}
