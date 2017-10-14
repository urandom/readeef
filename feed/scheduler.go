package feed

import (
	"bytes"
	"context"
	"crypto/md5"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/pool"
)

type Scheduler struct {
	ops    chan feedOp
	client *http.Client
	log    log.Log
}

type UpdateData struct {
	Feed    parser.Feed
	message string
}

func NewScheduler(log log.Log) Scheduler {
	return Scheduler{
		ops:    make(chan feedOp),
		client: &http.Client{Timeout: 30 * time.Second},
		log:    log,
	}
}

type feedOp func(feedMap)
type feedMap map[content.FeedID]schedulePayload

type schedulePayload struct {
	feed       content.Feed
	update     time.Duration
	updateData chan UpdateData
}

func (s Scheduler) ScheduleFeed(ctx context.Context, feed content.Feed, update time.Duration) <-chan UpdateData {
	ret := make(chan UpdateData)

	s.ops <- func(feedMap feedMap) {
		if _, ok := feedMap[feed.ID]; ok {
			return
		}

		payload := schedulePayload{
			feed:       feed,
			update:     update,
			updateData: ret,
		}
		feedMap[feed.ID] = payload

		go s.updateFeed(ctx, payload, []byte{})
	}

	return ret
}

func (s Scheduler) unscheduleFeed(ctx context.Context, feed content.Feed) {
	s.ops <- func(feedMap feedMap) {
		s.log.Infof("Unscheduling updates for feed %s", feed)
		payload := feedMap[feed.ID]
		close(payload.updateData)

		delete(feedMap, feed.ID)
	}
}

func (s Scheduler) updateFeed(ctx context.Context, payload schedulePayload, contentHash []byte) {
	select {
	case <-ctx.Done():
		s.unscheduleFeed(ctx, payload.feed)
		return
	default:
		var data UpdateData
		feed := payload.feed
		now := time.Now()

		if len(contentHash) == 0 || (!feed.SkipHours[now.Hour()] && !feed.SkipDays[now.Weekday().String()]) {
			data, contentHash = s.downloadFeed(payload, contentHash)
		}

		select {
		case <-ctx.Done():
			s.unscheduleFeed(ctx, payload.feed)
			return
		default:
			s.log.Debugf("Sending update data for feed %s", payload.feed)
			if data.isUpdated() || data.IsErr() {
				payload.updateData <- data
			}

			<-time.After(payload.update)
			s.updateFeed(ctx, payload, contentHash)
		}
	}
}

func (s Scheduler) downloadFeed(payload schedulePayload, contentHash []byte) (UpdateData, []byte) {
	feed := payload.feed

	s.log.Infof("Downloading content for feed %s", feed)
	resp, err := s.client.Get(feed.Link)

	if err != nil {
		return UpdateData{message: err.Error()}, contentHash
	} else if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		return UpdateData{message: "HTTP Status: " + strconv.Itoa(resp.StatusCode)}, contentHash
	} else {
		defer resp.Body.Close()

		buf := pool.Buffer.Get()
		defer pool.Buffer.Put(buf)

		if _, err := buf.ReadFrom(resp.Body); err == nil {
			hash := md5.Sum(buf.Bytes())
			if bytes.Equal(contentHash, hash[:]) {
				return UpdateData{}, contentHash
			}

			contentHash = hash[:]
			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				return UpdateData{Feed: pf}, contentHash
			} else {
				return UpdateData{message: err.Error()}, contentHash
			}
		} else {
			return UpdateData{message: err.Error()}, contentHash
		}
	}
}

func (u UpdateData) isUpdated() bool {
	return len(u.Feed.Articles) > 0 && !u.IsErr()
}

func (u UpdateData) IsErr() bool {
	return u.message != ""
}

func (u UpdateData) Error() string {
	return u.message
}

func (s Scheduler) Start(ctx context.Context) {
	feedMap := feedMap{}

	for {
		select {
		case op := <-s.ops:
			op(feedMap)
		case <-ctx.Done():
			return
		}
	}
}
