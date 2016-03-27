importScripts("/dist/moment/min/moment-with-locales.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        unreadIds = event.data.unreadIds || [],
        favoriteIds = event.data.favoriteIds || [],
        inserts = [], insertIndex = 0, cumulativeIndex = 0, lastUnreadIndex = -1,
        newArticles = event.data.newArticles,
        olderFirst = event.data.olderFirst,
        unreadOnly = event.data.unreadOnly,
        feeds = event.data.feeds,
        updateRequest = event.data.updateRequest,
        minId = 0, maxId = 0,
        articleMap = {}, indexMap = {}, stateChange = {}, unreadMap = {}, favoriteMap ={},
        feedMap, response;

    if (event.data.lang) {
        moment.locale(event.data.lang);
    }

    if (event.data.treatAsUnread) {
        articles[event.data.treatAsUnread].Read = false;
    }

    if (updateRequest) {
        for (var i = 0, id; id = unreadIds[i]; ++i) {
            unreadMap[id] = true;
        }

        for (var i = 0, id; id = favoriteIds[i]; ++i) {
            favoriteMap[id] = true;
        }
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;

        if (updateRequest) {
            var changes = {};

            if (unreadMap[a.Id]) {
                if (a.Read) {
                    a.Read = false;
                    changes.Read = false;
                }
            } else if (!a.Read) {
                a.Read = true;
                changes.Read = true;
            }

            if (favoriteMap[a.Id]) {
                if (!a.Favorite) {
                    a.Favorite = true;
                    changes.Favorite = true;
                }
            } else if (a.Favorite) {
                a.Favorite = false;
                changes.Favorite = false;
            }

            if (Object.keys(changes).length) {
                stateChange[a.Id] = changes;
            }
        }

        if (!a.Read) {
            lastUnreadIndex = i;
        }

        if (minId == 0 || a.Id < minId) {
            minId = a.Id
        }

        if (a.Id > maxId) {
            maxId = a.Id
        }
    }

    for (var i = 0, a, o, pre; a = newArticles[i]; ++i) {
        if (articleMap[a.Id]) {
            continue;
        }
		articleMap[a.Id] = a

        if (!unreadOnly || !a.Read) {
            if (feeds && feeds.length) {
                if (!feedMap) {
                    feedMap = {};
                    for (var j = 0, f; f = feeds[j]; ++j) {
                        feedMap[f.Id] = f;
                    }
                }

                var feed = feedMap[a.FeedId];
                if (!feed) {
                    continue;
                }
                a.FeedOrigin = feedMap[a.FeedId].Title;
            }

            a.ShortDescription = a.Description
                .replace(/<\w[^>]*>/g, '').replace(/<\/[^>]*>/g, '').trim().replace(/\s\s+/g, ' ');

            a.Date = new Date(a.Date);
            a.RelativeDate = moment(a.Date).fromNow();

            for (var o; o = articles[insertIndex]; ++insertIndex, ++cumulativeIndex) {
                // Unread articles are always first
                if (!a.Read && insertIndex > lastUnreadIndex) {
                    break;
                }

                if (a.Read == o.Read) {
                    if (olderFirst) {
                        if (o.Date > a.Date) {
                            break;
                        }
                    } else {
                        if (o.Date <= a.Date) {
                            break;
                        }
                    }
                }
            }

            if (!inserts[inserts.length - 1] || inserts[inserts.length - 1].index != cumulativeIndex - inserts[inserts.length - 1].articles.length) {
                inserts.push({index: cumulativeIndex, articles: []});
            }

            inserts[inserts.length - 1].articles.push(a);
            ++cumulativeIndex;

            if (minId == 0 || a.Id < minId) {
                minId = a.Id
            }

            if (a.Id > maxId) {
                maxId = a.Id
            }
        }
    }

    for (var i = 0, insert; insert = inserts[i]; ++i) {
        articles.splice.apply(articles, [insert.index, 0].concat(insert.articles));
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        indexMap[a.Id] = i;
    }

    response = {
        inserts: inserts,
        indexMap: indexMap,
        stateChange: stateChange,
        minId: minId,
        maxId: maxId,
        updateRequest: updateRequest,
    };
    if ('requestedArticle' in event.data) {
        response.requestedArticle = event.data.requestedArticle;
    }
    self.postMessage(response);
});
