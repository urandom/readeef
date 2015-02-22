package content

import (
	"fmt"

	"github.com/urandom/readeef/content/info"
)

type Subscription interface {
	Error

	fmt.Stringer

	Set(info info.Subscription)
	Info() info.Subscription

	Validate()

	Update(info info.Subscription)
	Delete()

	Fail(fail bool)
}
