(function() {
    "use strict";

    var userTTL = 1000 * 60 * 60 * 24 * 15,
        state = {VALIDATING: 1 << 0};

    Polymer({
        is: "rf-router",
        properties: {
            route: {
                type: String
            },
            user: {
                type: Object,
                readOnly: true
            }
        },
        routeNameMap: {},
        state: 0,

        attached: function() {
            if (!Object.keys(this.routeNameMap).length) {
                this._generateRouteMap();
            }

            if (!this.user && this.state & state | state.VALIDATING) {
                if (!MoreRouting.getRouteByName('splash').children[0].active) {
                    MoreRouting.navigateTo('login');
                } else if (!MoreRouting.getRouteByName('login').active &&
                    !MoreRouting.getRouteByName('login-from').active) {
                    MoreRouting.navigateTo('login-from', {url: encodeURIComponent(location.pathname)});
                }
            }
        },

        onRouteChange: function(event, detail) {
            if (!Object.keys(this.routeNameMap).length) {
                this._generateRouteMap();
            }

            switch (this.routeNameMap[detail.newRoute]) {
                case "login":
                case "login-from":
                    if (!this.user & this.state | state.VALIDATING) {
                        this.$.splash.selected = 0;
                    }
                    break;
            }
        },

        onUserLoad: function(event) {
            var storage = event.target;

            if (storage.value) {
                if (!storage.value.authTime || new Date().getTime() - storage.value.authTime > this.userTTL) {
                    storage.value = null;
                }
            }

            this.validateUser(storage.value);
        },

        validateUser: function(user) {
            this.state |= state.VALIDATING;
            this.async(function() {
                this.state &= ~state.VALIDATING;
            }.bind(this));

            var authCheck = this.$['auth-check'];
            var validateMessage = function(event, data) {
                console.log(data);
                this._setUser(user);

                if (MoreRouting.getRouteByName('login').active) {
                    Polymer.dom(this.root).querySelector('rf-login').hide();
                } else if (MoreRouting.getRouteByName('login-from').active) {
                    url = MoreRouting.getRouteByName('login-from').params.url;
                    Polymer.dom(this.root).querySelector('rf-login').hide();
                    MoreRouting.navigateTo(url);
                } else if (MoreRouting.getRouteByName('splash').active) {
                    MoreRouting.navigateTo('feed', {tag: 'all'});
                }
                this.$.splash.selected = 0;
                // TODO: test if user is valid
                authCheck.removeEventListener('rf-api-message', validateMessage);
            }.bind(this);

            authCheck.user = user;
            authCheck.addEventListener('rf-api-message', validateMessage);
            authCheck.send();
        },

        _generateRouteMap: function() {
            var map = {};
            ["splash", "login", "login-from", "feed-base", "settings-base"].forEach(function(n) {
                map[MoreRouting.getRouteByName(n)] = n;
            });
            this.routeNameMap = map;
        }

    })
})();
