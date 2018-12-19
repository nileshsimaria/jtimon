package main

import (
	"fmt"

	"google.golang.org/grpc"
)

var vendors = []*vendor{newJuniperJUNOS(), newCiscoIOSXR()}

type vendor struct {
	name               string
	loginCheckRequired bool
	dialExt            func(*JCtx) grpc.DialOption
	subscribe          func(*grpc.ClientConn, *JCtx, chan<- bool) SubErrorCode
}

func getVendor(jctx *JCtx) (*vendor, error) {
	name := jctx.config.Vendor.Name
	// juniper-junos is default
	if name == "" {
		name = "juniper-junos"
	}
	for _, vendor := range vendors {
		if name == vendor.name {
			return vendor, nil
		}
	}
	return nil, fmt.Errorf("The support for vendor [%s] has not implemented yet", name)
}

func newJuniperJUNOS() *vendor {
	return &vendor{
		name:               "juniper-junos",
		loginCheckRequired: true,
		dialExt:            nil,
		subscribe:          subscribeJuniperJUNOS,
	}
}

func newCiscoIOSXR() *vendor {
	return &vendor{
		name:               "cisco-iosxr",
		loginCheckRequired: false,
		dialExt:            getXRDialExtension,
		subscribe:          subscribeCiscoIOSXR,
	}
}
