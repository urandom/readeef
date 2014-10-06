(function() {
    "use strict";

    Polymer('rf-scaffolding', {
        display: "feed",
        articleRead: false,
        readStateJob: null,
        pathObserver: null,
        searchVisible: false,

        ready: function() {
            var drawerPanel = this.$['drawer-panel'];

            this.searchEnabled = this.searchEnabled == "true";

            this.$.navicon.addEventListener('click', function() {
                drawerPanel.togglePanel();
            });

            this.$['drawer-menu'].addEventListener('tap', function() {
                drawerPanel.togglePanel();
            });

            document.addEventListener('keypress', this.onContentKeypress.bind(this), false);
        },

        articleChanged: function() {
            var processArticleState = (function processArticleState(newArticle) {
                if (this.readStateJob) {
                    this.readStateJob.stop();
                }

                if (this.article.Read && !newArticle && this.articleIsRead) {
                    this.readStateJob = Polymer.job.call(this, this.readStateJob, function() {
                        this.articleRead = true;
                    }, 1000);
                } else {
                    this.articleRead = this.article.Read;
                }

                if (newArticle && !this.article.Read) {
                    this.articleIsRead = true;
                } else {
                    this.articleIsRead = false;
                }
            }).bind(this);

            if (this.pathObserver) {
                this.pathObserver.close();
            }

            if (this.article) {
                processArticleState(true);

                this.pathObserver = new PathObserver(this, 'article.Read');
                this.pathObserver.open(function() { processArticleState(false) });
            }

            this.$['drawer-panel'].disableSwipe = !!this.article;
        },

        onRefresh: function() {
            this.fire('core-signal', {name: "rf-feed-refresh"});
        },

        onArticleBack: function(event) {
            event.stopPropagation();
            event.preventDefault();

            if (this.display == 'feed') {
                this.async(function() {
                    this.article = null;
                });
            } else {
                this.display = 'feed';
            }
        },

        onArticleReadToggle: function() {
            this.fire('core-signal', {name: 'rf-read-article-toggle'});
        },

        onArticlePrevious: function() {
            this.fire('core-signal', {name: 'rf-previous-article'});
        },

        onArticleNext: function() {
            this.fire('core-signal', {name: 'rf-next-article'});
        },

        onOlderFirst: function() {
            this.settings.newerFirst = false;
        },

        onNewerFirst: function() {
            this.settings.newerFirst = true;
        },

        onUnreadOnly: function() {
            this.settings.unreadOnly = true;
        },

        onReadAndUnread: function() {
            this.settings.unreadOnly = false;
        },

        onMarkAllAsRead: function() {
            this.fire('core-signal', {name: 'rf-mark-all-as-read'});
        },

        onSearchToggle: function(event, detail, sender) {
            this.searchVisible = !this.searchVisible;
            if (this.searchVisible) {
                this.async(function() {
                    this.$['drawer-panel'].querySelector('.search-input').focus();
                });
            } else {
                if (sender && sender.classList.contains('search-close')) {
                    this.fire('core-signal', {name: 'rf-feed-search', data: ''});
                }
            }
        },

        onSearchKeyUp: function(event, detail, sender) {
            switch (event.keyCode) {
            case 13: //Enter
                this.fire('core-signal', {name: 'rf-feed-search', data: sender.value});
                this.onSearchToggle();
                break;
            case 27: //Escape
                this.onSearchToggle();
                break;
            }
        },

        onSearchKeyPress: function(event, detail, sender) {
            event.stopPropagation();
        },

        onContentKeypress: function(event) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (this.searchEnabled && (key == "Help" || key == "U+003F" || code == 47 || code == 63)) { // "/"
                if (!this.article && !this.searchVisible) {
                    this.onSearchToggle();
                }
            }
        }
    });
})();
