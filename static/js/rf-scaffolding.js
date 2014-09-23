(function() {
    "use strict";

    Polymer('rf-scaffolding', {
        display: "feed",

        ready: function() {
            var drawerPanel = this.$['drawer-panel'];

            this.$.navicon.addEventListener('click', function() {
                drawerPanel.togglePanel();
            });

            this.$['drawer-menu'].addEventListener('tap', function() {
                drawerPanel.togglePanel();
            });
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

    });
})();
