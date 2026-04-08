//go:build with_purego

package cronet

import "github.com/dkrhodes/cronet-go/internal/cronet"

func LoadLibrary(path string) error {
	return cronet.LoadLibrary(path)
}
