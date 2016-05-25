package controllers

import (
	"CatchARide-API/config"
)

type Response struct {
	Code    int64
	Error   string
	ErrorOn string
}

var globalConfig config.GlobalConfig
var envConfig config.EnvConfig

func RegConfig(gcfg config.GlobalConfig, envcfg config.EnvConfig) {
	globalConfig = gcfg
	envConfig = envcfg
}
