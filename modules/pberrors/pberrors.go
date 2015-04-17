package pberrors

import "encoding/json"
import (
//"errors"
)

type PbError struct {
	ErrorCode int    `json:"error_code,string"`
	ErrorMsg  string `json:"error_msg"`
}

func NewPbError(code int, msg string) *PbError {
	if code == 0 {
		return nil
	}
	return &PbError{code, msg}
}

func ExtractError(resp []byte) (error, *PbError) {
	var pberror PbError
	err := json.Unmarshal(resp, &pberror)
	if err != nil {
		return err, nil
	}
	if pberror.ErrorCode != 0 {
		return nil, &pberror
	}
	return nil, nil
}
