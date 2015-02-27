package readeef

import "github.com/urandom/readeef/content"

type UpdateFeedReceiver interface {
	UpdateFeedChannel() chan<- content.Feed
}

type UpdateFeedReceiverManager struct {
	updateReceivers []chan<- content.Feed
}

func (u *UpdateFeedReceiverManager) AddUpdateReceiver(r UpdateFeedReceiver) {
	ch := r.UpdateFeedChannel()
	for i := range u.updateReceivers {
		if u.updateReceivers[i] == ch {
			return
		}
	}

	u.updateReceivers = append(u.updateReceivers, ch)
}

func (u *UpdateFeedReceiverManager) RemoveUpdateReceiver(r UpdateFeedReceiver) {
	ch := r.UpdateFeedChannel()

	for i := range u.updateReceivers {
		if u.updateReceivers[i] == ch {
			u.updateReceivers = append(u.updateReceivers[:i], u.updateReceivers[i+1:]...)

			break
		}
	}
}

func (u UpdateFeedReceiverManager) NotifyReceivers(f content.Feed) {
	for i := range u.updateReceivers {
		u.updateReceivers[i] <- f
	}
}
