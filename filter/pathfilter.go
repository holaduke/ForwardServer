package filter

import (
	"strings"
)

func IsDisallowedReqPath(reqPath string) bool {

	var isStaticFile = false
	for _, v := range DefaultDisallowPath {
		if strings.Contains(strings.ToLower(reqPath), strings.ToLower(v)) {
			isStaticFile = true
			break
		}
	}

	return isStaticFile
}
