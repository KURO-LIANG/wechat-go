package context

import (
	"github.com/kuro-liang/wechat-go/credential"
	"github.com/kuro-liang/wechat-go/work/config"
)

// Context struct
type Context struct {
	*config.Config
	credential.AccessTokenHandle
}
