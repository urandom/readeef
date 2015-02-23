package content

import (
	"fmt"

	"github.com/urandom/readeef/content/info"
)

type Subscription interface {
	Error

	fmt.Stringer

	Set(info info.Subscription) Subscription
	Info() info.Subscription

	Validate() error

	Update(info info.Subscription) Subscription
	Delete() Subscription

	Fail(fail bool) Subscription
}
