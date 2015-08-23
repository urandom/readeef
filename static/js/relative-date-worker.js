importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";
    var dates = {}, articles = event.data.articles;

    for (var i = 0, a; a = articles[i]; ++i) {
        dates[a.Id] = moment(a.Date).fromNow();
    }

    self.postMessage({dates: dates, tagOrId: event.data.tagOrId});
});
