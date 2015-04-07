(function() {
    "use strict";

    Polymer('rf-content-list', {
        itemHeight: 48,
        listHeight: 0,
        pathObserver: null,
        listPosition: 0,
        articleRead: false,
        
        ready: function() {
            document.addEventListener('keypress', this.onContentKeypress.bind(this), false);
        },

        onUpdateRelativeDate: function() {
            window.requestAnimationFrame(function() {
                if (this.feed && this.articles && this.articles.length) {
                    var worker = new Worker('/js/relative-date-worker.js');

                    worker.addEventListener('message', function(event) {
                        if (this.feed.Id == event.data.feedId) {

                            var dates = event.data.dates;

                            for (var i = 0, a, d; a = this.articles[i]; ++i) {
                                d = dates[a.Id];
                                if (d) {
                                    a.RelativeDate = d;
                                }
                            }
                        }

                        this.job('relativeDateWorker', function() {
                            this.onUpdateRelativeDate();
                        }, 60000);
                    }.bind(this));

                    worker.postMessage({current: this.feed});
                } else {
                    this.job('relativeDateWorker', function() {
                        this.onUpdateRelativeDate();
                    }, 60000);
                }
            }.bind(this));
        },

        nextArticle: function(unread) {
            if (this.article) {
                var index = this.articles.indexOf(this.article),
                    article;

                if (unread) {
                    while (article = this.articles[++index]) {
                        if (!article.Read) {
                            break;
                        }
                    }
                } else {
                    article = this.articles[index + 1];
                }

                if (article) {
                    this.article = article;
                }
            } else {
                if (unread) {
                    var article, index = -1;
                    while (article = this.articles[++index]) {
                        if (!article.Read) {
                            break;
                        }
                    }

                    this.article = article;
                } else {
                    this.article = this.articles[0];
                }
            }
        },

        previousArticle: function(unread) {
            if (this.article) {
                var index = this.articles.indexOf(this.article),
                    article;

                if (unread) {
                    while (article = this.articles[--index]) {
                        if (!article.Read) {
                            break;
                        }
                    }
                } else if (index > 0) {
                    article = this.articles[index - 1];
                }

                if (article) {
                    this.article = article;
                }
            } else {
                if (unread) {
                    var article, index = this.articles.length;
                    while (article = this.articles[--index]) {
                        if (!article.Read) {
                            break;
                        }
                    }

                    this.article = article;
                } else {
                    this.article = this.articles[this.articles.length - 1];
                }
            }
        },

        created: function() {
            this.articles = [];
        },

        domReady: function() {
            this.$['articles-list'].scrollTarget =
                document.querySelector('rf-app').$.scaffolding.$['content-panel'].$.mainContainer;
        },

        feedChanged: function(oldValue, newValue) {
            this.listHeight = 0;

            this.article = null;

            var processArticles = (function processArticles() {
                if (this.feed.Articles && this.feed.Articles.length) {
                    var worker = new Worker('/js/content-articles-worker.js'),
                        data = {current: this.feed},
                        feedId = this.feed.Id.toString();

                    worker.addEventListener('message', function(event) {
                        this.articles = event.data.articles;
                        this.job('relativeDateWorker', function() {
                            this.onUpdateRelativeDate();
                        }, 60000);
                    }.bind(this));

                    if (feedId.indexOf("tag:") == 0 || feedId.indexOf("search:") == 0 || feedId.indexOf("popular:") == 0 || feedId == 'favorite' || feedId == 'all') {
                        data.feeds = this.feeds;
                    }

                    worker.postMessage(data);
                } else if (this.articles.length) {
                    this.articles = [];
                }
            }).bind(this);

            if (newValue) {
                processArticles();

                if (this.pathObserver) {
                    this.pathObserver.close();
                }

                this.pathObserver = new PathObserver(newValue, 'Articles');
                this.pathObserver.open(function(newValue) {
                    processArticles();
                }.bind(this));
            }
        },

        articleChanged: function(oldValue, newValue) {
            var scroller = document.querySelector('rf-app').$.scaffolding.$['content-panel'].scroller;

            if (newValue) {
                if (!oldValue) {
                    this.listPosition = scroller.scrollTop;
                }
                scroller.scrollTop = 0;

                this.$.pages.selected = 'detail';

                if (!newValue.Read) {
                    this.articleRead = true;
                    this.$['article-read'].send();
                }

                if (newValue.Last) {
                    this.asyncFire('core-signal', {name: 'rf-request-articles'});
                }
            } else {
                this.$.pages.selected = 'list';

                this.async(function() {
                    this.$['articles-list'].refresh(true);
                    scroller.scrollTop = this.listPosition;
                });
            }
        },

        articlesChanged: function(oldValue, newValue) {
            if (newValue && this.article && newValue.indexOf(this.article) == -1) {
                var found = false;
                for (var i = 0, a; a = newValue[i]; ++i) {
                    if (a.Id == this.article.Id) {
                        this.article = a;
                        found = true;
                        break;
                    }
                }

                if (!found) {
                    this.article = null;
                }
            }
        },

        onArticleActivate: function(event, detail) {
            this.article = detail.data;
        },

        onArticleNext: function(event) {
            this.nextArticle();
        },

        onArticlePrevious: function(event) {
            this.previousArticle();
        },

        onScrollThresholdTrigger: function(event) {
            if (!this.articles.length || this.$.pages.selected == "detail") {
                return;
            }

            this.asyncFire('core-signal', {name: 'rf-request-articles'});
        },

        onFavoriteToggle: function(event) {
            event.stopPropagation();
            event.preventDefault();

            var articleId = event.target.getAttribute('data-article-id');
            for (var i = 0, a; a = this.articles[i]; ++i) {
                if (a.Id == articleId) {
                    a.Favorite = !a.Favorite;

                    this.$['article-favorite'].send({id: a.Id, value: a.Favorite});
                    this.$['articles-list'].refresh(true);

                    break;
                }
            }
        },

        onReadArticleToggle: function() {
            this.articleRead = !this.article.Read;
            this.$['article-read'].send();
            this.$['articles-list'].refresh(true);
        },

        onArticleReadMessage: function(event, data) {
            if (this.article && this.article.Id == data.arguments.Id) {
                this.article.Read = data.arguments.Read;
            } else {
                for (var i = 0, a; a = this.articles[i]; ++i) {
                    if (a.Id == data.arguments.Id) {
                        a.Read = data.arguments.Read;
                        break;
                    }
                }
            }
        },

        onContentKeypress: function(event) {
            if (this.$.pages.offsetWidth == 0 && this.$.pages.offsetHeight == 0) {
                return;
            }

            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == "U+004A" || code == 106 || code == 74) { // j
                this.nextArticle(event.shiftKey);
            } else if (key == "U+004B" || code == 107 || code == 75) { // k
                this.previousArticle(event.shiftKey);
            } else if (key == "U+0048" || code == 104 || code == 72) { // h
                this.article = null;
            } else if (key == "U+0056" || code == 118 || code == 86) { // v
                if (this.article) {
                    this.$['article-display'].openCurrentArticle();
                }
            } else if (key == "U+004D" || code == 109 || code == 77) { // m
                if (this.article) {
                    this.onReadArticleToggle();
                }
            } else if (key == "U+0043" || code == 99 || code == 67) { // c
                this.fire('core-signal', {name: "rf-article-format"});
            } else if (key == "U+0053" || code == 115 || code == 83) { // s
                this.fire('core-signal', {name: "rf-article-summarize"});
            } else if (key == "U+0052" || code == 114 || code == 82) { // r
                this.fire('core-signal', {name: "rf-feed-refresh"});
            } else if (key == "U+0046" || code == 102 || code == 70) { // f
                if (this.article) {
                    this.fire('core-signal', {name: "rf-article-favorite"});
                }
            } else if (key == "U+003F" || key == "Help" || code == "47" || code == "63") {
            }
        }
    });
})();
