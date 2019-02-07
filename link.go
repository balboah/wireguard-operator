package operator

import (
	"github.com/vishvananda/netlink"
)

type WgLink struct {
	netlink.LinkAttrs
}

func (WgLink) Type() string {
	return "wireguard"
}

func (l *WgLink) Attrs() *netlink.LinkAttrs {
	return &l.LinkAttrs
}
