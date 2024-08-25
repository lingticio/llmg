package utils

import (
	"os"
	"strings"
)

func IsInUnitTest() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}

	return false
}
