package ttrss

import (
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
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

func getApiLevel(req request, user content.User, service repo.Service) (interface{}, error) {
	return genericContent{Level: API_LEVEL}, nil
}

func getVersion(req request, user content.User, service repo.Service) (interface{}, error) {
	return genericContent{Version: API_VERSION}, nil
}

func getConfig(req request, user content.User, service repo.Service) (interface{}, error) {
	feeds, err := service.FeedRepo().ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user feeds")
	}
	feedCount := len(feeds)

	return configContent{DaemonIsRunning: true, NumFeeds: feedCount}, nil
}

func unknown(req request, user content.User, service repo.Service) (interface{}, error) {
	return genericContent{Method: req.Op}, errors.WithStack(newErr("unknown method "+req.Op, "UNKNOWN_METHOD"))
}

func init() {
	actions["getApiLevel"] = getApiLevel
	actions["getVersion"] = getVersion
	actions["getConfig"] = getConfig
}
