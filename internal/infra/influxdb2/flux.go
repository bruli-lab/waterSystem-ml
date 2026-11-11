package influxdb2

import (
	"strconv"
	"strings"
)

func fluxStringArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, v := range values {
		quoted = append(quoted, strconv.Quote(v))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
