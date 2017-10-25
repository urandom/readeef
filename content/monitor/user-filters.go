package monitor

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/log"
)

func UserFilters(service eventable.Service, log log.Log) {
	userRepo := service.UserRepo()

	for event := range service.Listener() {
		switch data := event.Data.(type) {
		case eventable.FeedSetTagsData:
			original := content.GetUserFilters(data.User)
			filters := make([]content.Filter, 0, len(original))
			tagIDs := map[content.TagID]struct{}{}

			for _, t := range data.Tags {
				tagIDs[t.ID] = struct{}{}
			}

			var changed bool
			for i := range original {
				if original[i].TagID > 0 {
					if _, ok := tagIDs[original[i].TagID]; ok {
						// Add the feed id to the list of feeds
						var found bool
						for _, id := range original[i].FeedIDs {
							if id == data.Feed.ID {
								found = true
								break
							}
						}

						if !found {
							original[i].FeedIDs = append(
								original[i].FeedIDs,
								data.Feed.ID,
							)
							changed = true
						}
					} else {
						// Remove the feed id, if it was in the list
						for j := range original[i].FeedIDs {
							if original[i].FeedIDs[j] == data.Feed.ID {
								original[i].FeedIDs = append(
									original[i].FeedIDs[:j],
									original[i].FeedIDs[j+1:]...,
								)

								changed = true
								break
							}
						}
					}
				}

				if original[i].Valid() {
					filters = append(filters, original[i])
				}
			}

			if len(filters) != len(original) || changed {
				data.User.ProfileData["filters"] = filters
				if err := userRepo.Update(data.User); err != nil {
					log.Printf("Error updating user %s: %+v", data.User, err)
				}
			}
		}
	}
}
