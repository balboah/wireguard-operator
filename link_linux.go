package operator

import (
	"github.com/vishvananda/netlink"
)

func (wg *WgLink) AddrAdd(addrs ...*netlink.Addr) error {
	for _, addr := range addrs {
		if err := netlink.AddrReplace(wg, addr); err != nil {
			return err
		}
	}
	return nil
}
