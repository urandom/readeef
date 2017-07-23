package ttrss

import (
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

type configContent struct {
	IconsDir        string `json:"icons_dir"`
	IconsUrl        string `json:"icons_url"`
	DaemonIsRunning bool   `json:"daemon_is_running"`
	NumFeeds        int    `json:"num_feeds"`
}

type genericContent struct {
	Level     int         `json:"level,omitempty"`
	ApiLevel  int         `json:"api_level,omitempty"`
	Version   string      `json:"version,omitempty"`
	SessionId string      `json:"session_id,omitempty"`
	Status    interface{} `json:"status,omitempty"`
	Unread    string      `json:"unread,omitempty"`
	Updated   int64       `json:"updated,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Method    string      `json:"method,omitempty"`
}

func getApiLevel(req request, user content.User) (interface{}, error) {
	return genericContent{Level: TTRSS_API_LEVEL}, nil
}

func getVersion(req request, user content.User) (interface{}, error) {
	return genericContent{Version: TTRSS_VERSION}, nil
}

func getConfig(req request, user content.User) (interface{}, error) {
	feedCount := len(user.AllFeeds())
	if user.HasErr() {
		return nil, errors.Wrapf(user.Err(), "getting user %s feeds", user.Data().Login)
	}

	return configContent{DaemonIsRunning: true, NumFeeds: feedCount}, nil
}

func unknown(req request, user content.User) (interface{}, error) {
	return genericContent{Method: req.Op}, errors.WithStack(newErr("unknown method "+req.Op, "UNKNOWN_METHOD"))
}

func init() {
	actions["getApiLevel"] = getApiLevel
	actions["getVersion"] = getVersion
	actions["getConfig"] = getConfig
}
