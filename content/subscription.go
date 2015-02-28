package content

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/readeef/content/info"
)

type Subscription interface {
	Error
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Info(in ...info.Subscription) info.Subscription

	Validate() error

	Update()
	Delete()
}
