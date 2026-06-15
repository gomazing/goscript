package buildout

import (
	"github.com/gomazing/goscript/pkg/hyper"
)

func writeHyperFile(path string, value interface{}) error {
	return hyper.WriteFile(path, value)
}

func mustMarshalHyper(value interface{}) []byte {
	data, err := hyper.MarshalIndent(value, "", "  ")
	if err != nil {
		return []byte("<hyper kind=\"error\"></hyper>")
	}
	return data
}
