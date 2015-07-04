package process

import "sort"
import "bytes"

import "purstal/zhidingshanlou/web"
import "purstal/zhidingshanlou/misc"
import pkg_account "purstal/zhidingshanlou/core/account"

func AddSignature(parameters *web.Parameters) {
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

func AddMandatoryParams(parameters *web.Parameters, account *pkg_account.Account) {
	if account.BDUSS != "" {
		parameters.Add(PARAM_BDUSS, account.BDUSS)
	}
	parameters.Add(PARAM_NET_TYPE, account.PUBLICnet_type)                      //3
	parameters.Add(PARAM_CLIENT_TYPE, account.PUBLIC_client_type)               //Consts.CLIENT_TYPE).ToString()
	parameters.Add(PARAM_CLIENT_ID, account.PUBLIC_client_id)                   //States.DeviceId
	parameters.Add(PARAM_CLIENT_VERSION, account.PUBLIC_client_version)         //States.Version
	parameters.Add(PARAM_PHONE_IMEI, misc.ComputeMD5(account.PUBLIC_client_id)) //ComputeMD5(States.DeviceId)
}

func ProcessParams(parameters *web.Parameters, account *pkg_account.Account) {
	if account != nil {
		AddMandatoryParams(parameters, account)
	}
	AddSignature(parameters)
}
