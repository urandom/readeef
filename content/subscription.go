package content

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/readeef/content/data"
)

type Subscription interface {
	Error
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Data(data ...data.Subscription) data.Subscription

	Validate() error

	Update()
	Delete()
}
