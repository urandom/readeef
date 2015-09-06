importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        inserts = [], insertIndex = 0, cumulativeIndex = 0,
        newArticles = event.data.newArticles,
        newerFirst = event.data.newerFirst,
        unreadOnly = event.data.unreadOnly,
        feeds = event.data.feeds,
        articleMap = {}, indexMap = {}, feedMap, response;

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;
    }

    for (var i = 0, a, pre; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id] && !(unreadOnly && a.Read)) {
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
                if (!a.Read && o.Read) {
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

    response = {inserts: inserts, indexMap: indexMap, requestedArticle: event.data.requestedArticle};
    if ('requestedArticle' in event.data) {
        response.requestedArticle = event.data.requestedArticle;
    }
    self.postMessage(response);
});
