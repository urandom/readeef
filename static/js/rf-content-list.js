(function() {
    "use strict";

    Polymer('rf-content-list', {
        itemHeight: 48,
        listHeight: 0,
        pathObserver: null,
        listPosition: 0,
        
        ready: function() {
            document.addEventListener('keypress', this.onContentKeypress.bind(this), false);
        },

        nextArticle: function() {
            if (this.article) {
                var index = this.articles.indexOf(this.article);

                if (index < this.articles.length - 1) {
                    this.article = this.articles[index + 1];
                }
            } else {
                this.article = this.articles[0];
            }
        },

        previousArticle: function() {
            if (this.article) {
                var index = this.articles.indexOf(this.article);

                if (index) {
                    this.article = this.articles[index - 1];
                }
            } else {
                this.article = this.articles[this.articles.length - 1];
            }
        },

        created: function() {
            this.articles = [];
        },

        domReady: function() {
            var contentPanel = document.querySelector('rf-app').$.scaffolding.$['content-panel'];

            this.$['articles-list'].scrollTarget = contentPanel;

            contentPanel.addEventListener('scroll', function(event) {
                if (!this.articles.length || this.$.pages.selected == "detail") {
                    return;
                }

                if (contentPanel.scroller.offsetHeight + contentPanel.scroller.scrollTop + 50 > contentPanel.scroller.scrollHeight) {
                    this.asyncFire('core-signal', {name: 'rf-request-articles'});
                }
            }.bind(this));
        },

        feedChanged: function(oldValue, newValue) {
            this.listHeight = 0;

            this.article = null;

            var processArticles = (function processArticles() {
                if (this.feed.Articles && this.feed.Articles.length) {
                    var worker = new Worker('/js/content-articles-worker.js'),
                        data = {current: this.feed};

                    worker.addEventListener('message', function(event) {
                        this.articles = event.data.articles;
                    }.bind(this));
                    if (this.feed.Id.toString().indexOf("tag:") == 0) {
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

                this.$.pages.selected = 'detail';

                if (!newValue.Read) {
                    newValue.Read = true;
                    this.$['article-read'].go();
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

        onFavoriteToggle: function(event) {
            event.stopPropagation();
            event.preventDefault();

            var articleId = event.target.getAttribute('data-article-id');
            for (var i = 0, a; a = this.articles[i]; ++i) {
                if (a.Id == articleId) {
                    a.Favorite = !a.Favorite;

                    var clone = this.$['article-favorite'].cloneNode(true);
                    clone.pathAction = "article/favorite/" + a.FeedId + "/" + a.Id + "/" + a.Favorite;
                    clone.user = this.user;

                    clone.go()

                    this.$['articles-list'].refresh(true);

                    break;
                }
            }
        },

        onFavoriteArticleToggle: function() {
            this.article.Favorite = !this.article.Favorite;
            this.$['article-favorite'].go();
            this.$['articles-list'].refresh(true);
        },

        onContentKeypress: function(event) {
            if (this.$.pages.offsetWidth == 0 && this.$.pages.offsetHeight == 0) {
                return;
            }

            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == "U+004A" || code == 106 || code == 74) { // j
                this.nextArticle();
            } else if (key == "U+004B" || code == 107 || code == 75) { // k
                this.previousArticle();
            } else if (key == "U+0048" || code == 104 || code == 72) { // h
                this.article = null;
            } else if (key == "U+0056" || code == 118 || code == 86) { // v
                if (this.article) {
                    this.$['article-display'].openCurrentArticle();
                }
            } else if (key == "U+0052" || code == 114 || code == 82) { // r
                this.fire('core-signal', {name: "rf-feed-refresh"});
            }
        }
    });
})();
