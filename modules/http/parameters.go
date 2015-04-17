package http

import "bytes"
import "strings"
import "net/url"

type KeyValuePair struct {
	Key   string
	Value string
}

type Parameters []KeyValuePair
type Cookies []KeyValuePair

func (parameters *Parameters) Add(key, value string) {
	*parameters = append(*parameters, KeyValuePair{key, value})
}

func (parameters *Cookies) Add(key, value string) {
	*parameters = append(*parameters, KeyValuePair{key, value})
}

func (parameters Parameters) Encode() string {
	var buffer bytes.Buffer
	for _, parameter := range parameters {
		//buffer.WriteString(url.QueryEscape(parameter.Key))
		buffer.WriteString(parameter.Key)
		buffer.WriteRune(sCharEqual)
		//buffer.WriteString(url.QueryEscape(parameter.Value))
		buffer.WriteString(parameter.Value)
		buffer.WriteRune(sCharAnd)
	}

	return strings.TrimRight(buffer.String(), "&")
}

func (parameters Cookies) Encode() string {
	var buffer bytes.Buffer
	for _, parameter := range parameters {
		buffer.WriteString(url.QueryEscape(parameter.Key))
		buffer.WriteRune(sCharEqual)
		buffer.WriteString(url.QueryEscape(parameter.Value))
		buffer.WriteString("; ")
	}

	return strings.TrimRight(buffer.String(), "&")
}
