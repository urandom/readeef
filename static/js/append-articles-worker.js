importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        unshift = [], push = [],
        newArticles = event.data.newArticles,
        unshiftFirst = event.data.unshiftFirst,
        feeds = event.data.feeds,
        articleMap = {}, indexMap = {}, feedMap;

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;
    }

    for (var i = 0, a, pre; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id]) {
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

            a.Description = a.Description.replace(/<!--.*?-->/g, '')
                .replace(/<script.*?<\/script\s*>/g, '')
                .replace(/(<a )(.*?<\/a\s*>)/g, '$1 target="feed-article" $2');

            pre = "";
            while (pre != a.Description) {
                if (pre) {
                    a.Description = pre;
                }

                pre = a.Description.replace(/(<\w[^>]*?["'])javascript:([^>]*>)/g, '$1$2')
                    .replace(/(<\w[^>]*?)\s+on\w+=(['"]).*?\2([^>]*>)/g, '$1$3');
            }

            a.ShortDescription = a.Description
                .replace(/<\w[^>]*>/g, '').replace(/<\/[^>]*>/g, '').trim().replace(/\s\s+/g, ' ');

            a.RelativeDate = moment(a.Date).fromNow();

            if (unshiftFirst) {
                unshift.unshift(a);
            } else {
                push.push(a);
            }
        } else {
            unshiftFirst = false;
        }
    }

    var cumul = 0;
    [unshift, articles, push].forEach(function(list) {
        for (var i = 0, a; a = list[i]; ++i, ++cumul) {
            indexMap[a.Id] = cumul;
        }
    });

    self.postMessage({push: push, unshift: unshift, indexMap: indexMap});
});
