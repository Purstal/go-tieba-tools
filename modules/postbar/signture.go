package postbar

import "sort"
import "bytes"

import "github.com/purstal/pbtools/modules/http"
import "github.com/purstal/pbtools/modules/misc"

func AddSignature(parameters *http.Parameters) {
	//AddSignature
	list := make([]string, len(*parameters))
	for i := range list {
		list[i] = (*parameters)[i].Key + string(sCharEqual) + (*parameters)[i].Value
	}
	sort.Strings(list)
	var buffer bytes.Buffer
	for _, str := range list {
		buffer.WriteString(str)
	}
	buffer.WriteString(SIGN_KEY)
	parameters.Add(PARAM_SIGN, misc.ComputeMD5(buffer.String()))
}

func AddMandatoryParams(parameters *http.Parameters, account *Account) {
	//AddMandatoryParams
	if account.BDUSS != "" {
		parameters.Add(PARAM_BDUSS, account.BDUSS)
	}
	if PARAM_NET_TYPE != "" {
		parameters.Add(PARAM_NET_TYPE, account.NetType) //3
	}
	parameters.Add(PARAM_CLIENT_TYPE, account.ClientType)       //Consts.CLIENT_TYPE).ToString()
	parameters.Add(PARAM_CLIENT_ID, account.ClientID)           //States.DeviceId
	parameters.Add(PARAM_CLIENT_VERSION, account.ClientVersion) //States.Version
	if account.PhoneIMEI != "" {
		parameters.Add(PARAM_PHONE_IMEI, account.PhoneIMEI)
	}
}

func ProcessParams(parameters *http.Parameters, account *Account) {
	//ProcessParams
	if account != nil {
		AddMandatoryParams(parameters, account)
	}
	AddSignature(parameters)
}

const (
	sCharEqual           = '='
	SIGN_KEY             = "tiebaclient!!!"
	PARAM_SIGN           = "sign"
	PARAM_BDUSS          = "BDUSS"
	PARAM_NET_TYPE       = "net_type"
	PARAM_CLIENT_ID      = "_client_id"
	PARAM_CLIENT_TYPE    = "_client_type"
	PARAM_CLIENT_VERSION = "_client_version"
	PARAM_PHONE_IMEI     = "_phone_imei"
)
