package config

import (
	"github.com/kuro-liang/wechat-go/cache"
)

// Config .config for 微信公众号
type Config struct {
	AppID                string `json:"app_id"`               // appid
	AppSecret            string `json:"app_secret"`           // appsecret
	Token                string `json:"token"`                // token
	EncodingAESKey       string `json:"encoding_aes_key"`     // EncodingAESKey
	UseStableAccessToken bool   `json:"useStableAccessToken"` // 是否使用稳定的access_token
	Cache                cache.Cache
}
