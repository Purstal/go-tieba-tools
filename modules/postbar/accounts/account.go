package accounts

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/pberrors"
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

func NewDefaultAndroidAccount(id string) *Account {
	return &Account{
		ID:            id,
		NetType:       ``,
		ClientType:    Android,
		ClientID:      ``,
		ClientVersion: `6.1.3`,
		PhoneIMEI:     misc.ComputeMD5(``), //...
	}

}

func NewDefaultWindows8Account(id string) *Account {
	return &Account{
		ID:         id,
		NetType:    `3`,
		ClientType: Windows8,
		ClientID:   `4C-07-16-00-F1-C0-5B-47-62-86-B7-35-AF-24-24-DB-E7-05-86-8B-BF-E6-A4-06-B2-54-E3-AB-81-2D-9D-32`,
		//Maribel Hearn ↑
		ClientVersion: `1.5.0.0`,
		PhoneIMEI:     misc.ComputeMD5(``),
	}
}

func (acc *Account) GetTbs() (string, error, *pberrors.PbError) {
	resp, err := GetTbs(acc)
	if err != nil {
		return "", err, nil
	}
	var x struct {
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		Tbs       string `json:"tbs"`
	}
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		return "", err, nil
	}
	if x.ErrorCode != 0 {
		return "", nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	return x.Tbs, nil, nil

}

func (acc *Account) Login(password string) (error, *pberrors.PbError) {
	//resp, err := APILogin(acc, acc.ID, password)
	resp, err := Login(acc, acc.ID, password)
	if err != nil {
		return err, nil
	}
	var x struct {
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		User      struct {
			BDUSS string `json:"BDUSS"`
		} `json:"user"`
	}

	err2 := json.Unmarshal(resp, &x)

	if err2 != nil {
		return err2, nil
	}
	if x.ErrorCode != 0 {
		return nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	acc.BDUSS = x.User.BDUSS
	return nil, nil
}

func IsLogin(BDUSS string) (bool, error) {
	resp, err := GetTbsWeb(BDUSS)
	if err != nil {
		return false, err
	}
	var x struct {
		Tbs     string `json:"tbs"`
		IsLogin int    `json:"is_login"`
	}
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		return false, err2
	}
	return x.IsLogin == 1, nil
}

func (acc *Account) IsLogin() (bool, error) {
	return IsLogin(acc.BDUSS)
}
