package context

import (
	"github.com/kuro-liang/wechat-go/credential"
	"github.com/kuro-liang/wechat-go/officialaccount/config"
)

// Context struct
type Context struct {
	*config.Config
	credential.AccessTokenHandle
}
