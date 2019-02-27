package operator

import (
	"github.com/vishvananda/netlink"
)

type WgLink struct {
	*netlink.LinkAttrs
}

// Use existing link or create a new one.
func NewWgLink(ifName string) (*WgLink, error) {
	attr := netlink.NewLinkAttrs()
	attr.Name = ifName
	wg := WgLink{LinkAttrs: &attr}
	l, err := netlink.LinkByName(ifName)
	if err != nil {
		switch err.(type) {
		case netlink.LinkNotFoundError:
			if err := netlink.LinkAdd(&wg); err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	} else {
		wg.LinkAttrs = l.Attrs()
	}
	return &wg, netlink.LinkSetUp(&wg)
}

func (WgLink) Type() string {
	return "wireguard"
}

func (wg *WgLink) Attrs() *netlink.LinkAttrs {
	return wg.LinkAttrs
}

func (wg *WgLink) Close() error {
	return netlink.LinkDel(wg)
}
