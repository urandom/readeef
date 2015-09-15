importScripts("/dist/moment/min/moment-with-locales.min.js");

self.addEventListener('message', function(event) {
    "use strict";
    var dates = {}, articles = event.data.articles;

    if (event.data.lang) {
        moment.locale(event.data.lang);
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        dates[a.Id] = moment(a.Date).fromNow();
    }

    self.postMessage({dates: dates, tagOrId: event.data.tagOrId});
});
