package api

import (
	"github.com/json-iterator/go"
	"github.com/xin053/ipd/config"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// API ip api interface
type API interface {
	// Name get ip api name
	Name() string

	//Url get ip api url
	Url() string

	//Request get ip api result
	Request(ip string) (API, error)

	//JSON IPInfo
	JSON() config.IPInfo
}
