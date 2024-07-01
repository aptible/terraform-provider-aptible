package aptible

import (
	"strconv"
	"strings"
)

func ExtractIdFromLink(relation string) int32 {
	if relation == "" {
		return 0
	}
	segments := strings.Split(relation, "/")
	if len(segments) == 0 {
		return 0
	}
	val, _ := strconv.Atoi(segments[len(segments)-1])
	return int32(val)
}
