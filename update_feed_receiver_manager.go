package readeef

type UpdateFeedReceiver interface {
	UpdateFeedChannel() chan<- Feed
}

type UpdateFeedReceiverManager struct {
	updateReceivers []chan<- Feed
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

func (u UpdateFeedReceiverManager) NotifyReceivers(f Feed) {
	Debug.Printf("Notifying %d receivers of a feed update\n", len(u.updateReceivers))
	for i := range u.updateReceivers {
		u.updateReceivers[i] <- f
	}
}
