package dgraph

import (
	"fmt"
	"strconv"
)

type Uid struct {
	Uid string `json:"uid"`
}

func NewUid(id int64) Uid {
	return Uid{Uid: intToUid(id)}
}

func (u Uid) ToInt() int64 {
	return uidToInt(u.Uid)
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
