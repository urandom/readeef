importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";
    var dates = {}, current = event.data.current;

    for (var i = 0, a; a = current.Articles[i]; ++i) {
        dates[a.Id] = moment(a.Date).fromNow();
    }

    self.postMessage({dates: dates, feedId: current.Id});
});
