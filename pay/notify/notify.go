package notify

import (
	"github.com/kuro-liang/wechat/pay/config"
)

// Notify 回调
type Notify struct {
	*config.Config
}

// NewNotify new
func NewNotify(cfg *config.Config) *Notify {
	return &Notify{cfg}
}
