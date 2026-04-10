//go:build darwin && !ios && arm64

package all

import (
	_ "github.com/dkrhodes/cronet-go"
	_ "github.com/dkrhodes/cronet-go/lib/darwin_arm64"
)
