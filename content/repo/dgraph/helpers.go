package dgraph

import (
	"fmt"
	"strconv"
	"strings"
)

type UID struct {
	Value string `json:"uid"`
}

func NewUID(id int64) UID {
	return UID{Value: intToUid(id)}
}

func (u UID) ToInt() int64 {
	return uidToInt(u.Value)
}

func (u UID) Valid() bool {
	return u.Value != ""
}

func uidToInt(uid string) int64 {
	i, err := strconv.ParseInt(uid[2:], 16, 64)
	if err != nil {
		panic(err)
	}

	return i
}

func intToUid(id int64) string {
	return fmt.Sprintf("0x%s", strconv.FormatInt(id, 16))
}

func aliasPredicates(predicates string) string {
	fields := strings.Fields(predicates)

	aliased := make([]string, len(fields)*2)

	for i, j := 0, 0; i < len(fields); i, j = i+1, j+2 {
		aliased[j] = fields[i] + ":"
		aliased[j+1] = fields[i]
	}

	return strings.Join(aliased, " ")
}
