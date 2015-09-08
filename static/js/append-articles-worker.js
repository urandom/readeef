importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        inserts = [], insertIndex = 0, cumulativeIndex = 0, lastUnreadIndex = -1,
        newArticles = event.data.newArticles,
        newerFirst = event.data.newerFirst,
        unreadOnly = event.data.unreadOnly,
        feeds = event.data.feeds,
        articleMap = {}, indexMap = {}, stateChange = {},
        feedMap, response;

    if (event.data.treatAsUnread) {
        articles[event.data.treatAsUnread].Read = false;
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;
        if (!a.Read) {
            lastUnreadIndex = i;
        }
    }

    for (var i = 0, a, o, pre; a = newArticles[i]; ++i) {
        o = articleMap[a.Id];
        if (o) {
            if (o.Read != a.Read) {
                stateChange[a.Id] = {Read: a.Read};
            }

            if (o.Favorite != a.Favorite) {
                if (stateChange[a.Id]) {
                    stateChange[a.Id].Favorite = a.Favorite;
                } else {
                    stateChange[a.Id] = {Favorite: a.Favorite};
                }
            }
        } else if (!unreadOnly || !a.Read) {
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
                    if (newerFirst) {
                        if (o.Date <= a.Date) {
                            break;
                        }
                    } else {
                        if (o.Date > a.Date) {
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
        }
    }

    for (var i = 0, insert; insert = inserts[i]; ++i) {
        articles.splice.apply(articles, [insert.index, 0].concat(insert.articles));
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        indexMap[a.Id] = i;
    }

    response = {inserts: inserts, indexMap: indexMap, stateChange: stateChange};
    if ('requestedArticle' in event.data) {
        response.requestedArticle = event.data.requestedArticle;
    }
    self.postMessage(response);
});
