//go:build linux && !android && amd64 && !with_musl

package all

import (
	_ "github.com/dkrhodes/cronet-go"
	_ "github.com/dkrhodes/cronet-go/lib/linux_amd64"
)
