package context

import (
	"github.com/kuro-liang/wechat/credential"
	"github.com/kuro-liang/wechat/work/config"
)

// Context struct
type Context struct {
	*config.Config
	credential.AccessTokenHandle
}
