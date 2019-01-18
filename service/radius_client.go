package service

import (
	"context"
	"fmt"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type RadiusClient struct {
	host   string
	port   int
	secret string
}

func (r RadiusClient) Login(user, pass string) (code radius.Code, err error) {
	packet := radius.New(radius.CodeAccessRequest, []byte(r.secret))
	rfc2865.UserName_SetString(packet, user)
	rfc2865.UserPassword_SetString(packet, pass)
	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", r.host, r.port))
	if err != nil {
		code = radius.CodeAccessReject
	} else {
		code = response.Code
	}
	return code, err
}

func (r RadiusClient) Logout(user, pass string) (code radius.Code, err error) {
	packet := radius.New(radius.CodeDisconnectRequest, []byte(r.secret))
	rfc2865.UserName_SetString(packet, user)
	rfc2865.UserPassword_SetString(packet, pass)
	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", r.host, r.port))
	if err != nil {
		code = radius.CodeAccessReject
	} else {
		code = response.Code
	}
	return code, err
}
