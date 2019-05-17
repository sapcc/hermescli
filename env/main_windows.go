package env

import (
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
)

func Get(s string) string {
	env := os.Getenv(s)
	if windows.GetACP() == 1252 {
		if v, err := charmap.Windows1252.NewEncoder().String(env); err == nil {
			return v
		}
	}
	return env
}
