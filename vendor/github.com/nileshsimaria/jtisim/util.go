package jtisim

import (
	"time"
)

// MakeMSTimestamp timestamp in ms for JTI
func MakeMSTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
