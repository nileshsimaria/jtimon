package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

var vendors = []*vendor{newGNMI(), newJuniperJUNOS(), newCiscoIOSXR()}

type vendor struct {
	name               string
	loginCheckRequired bool
	sendLoginCheck     func(*JCtx, *grpc.ClientConn) error
	dialExt            func(*JCtx) grpc.DialOption
	subscribe          func(*grpc.ClientConn, *JCtx) SubErrorCode
	udpSubscribe       func(net.PacketConn, *JCtx) SubErrorCode
}

func getVendor(jctx *JCtx, tryGnmi bool) (*vendor, error) {
	name := jctx.config.Vendor.Name

	if tryGnmi {
		name = "gnmi"
	}
	// juniper-junos is default
	if name == "" {
		name = "juniper-junos"
	}
	for _, vendor := range vendors {
		if name == vendor.name {
			return vendor, nil
		}
	}
	return nil, fmt.Errorf("support for vendor [%s] has not implemented yet", name)
}

func newJuniperJUNOS() *vendor {
	return &vendor{
		name:               "juniper-junos",
		loginCheckRequired: true,
		sendLoginCheck:     loginCheckJunos,
		dialExt:            nil,
		subscribe:          subscribeJunos,
		udpSubscribe:       udpSubscribeJunos,
	}
}

func newCiscoIOSXR() *vendor {
	return &vendor{
		name:               "cisco-iosxr",
		loginCheckRequired: false,
		sendLoginCheck:     nil,
		dialExt:            dialExtensionXR,
		subscribe:          subscribeXR,
	}
}

func newGNMI() *vendor {
	return &vendor{
		name:               "gnmi",
		loginCheckRequired: false,
		sendLoginCheck:     nil,
		dialExt:            nil,
		subscribe:          subscribegNMI,
	}
}
