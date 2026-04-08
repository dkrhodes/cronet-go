//go:build with_purego

package cronet

import "github.com/dkrhodes/cronet-go/internal/cronet"

func checkLibrary() error {
	return cronet.LoadLibrary("")
}
