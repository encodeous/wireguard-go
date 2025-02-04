//go:build !linux

package device

import (
	"github.com/encodeous/wireguard-go/conn"
	"github.com/encodeous/wireguard-go/rwcancel"
)

func (device *Device) startRouteListener(bind conn.Bind) (*rwcancel.RWCancel, error) {
	return nil, nil
}
