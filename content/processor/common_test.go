package processor

import (
	"os"

	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/log"
)

var (
	logger log.Log
)

func init() {
	cfg := config.Log{}
	cfg.Converted.Writer = os.Stderr
	cfg.Converted.Prefix = "[testing] "
	logger = log.WithStd(cfg)
}
