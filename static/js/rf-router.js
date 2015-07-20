(function() {
    "use strict";

    var userTTL = 1000 * 60 * 60 * 24 * 15;

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

        attached: function() {
            if (!Object.keys(this.routeNameMap).length) {
                this._generateRouteMap();
            }
        },

        onRouteChange: function(event, detail) {
            if (!Object.keys(this.routeNameMap).length) {
                this._generateRouteMap();
            }

            switch (this.routeNameMap[detail.newRoute]) {
                case "splash":
                case "feed-base":
                case "settings-base":
                /*
                    if (this.user) {
                        this.validateUser(this.user);
                    } else {
                        MoreRouting.navigateTo('login');
                    }
                    */
                    break;
                case "login":
                case "login-from":
                    break;
            }
        },

        validateUser: function(user) {
            Polymer.dom(this.root).querySelector('rf-login').hide();
            // TODO: test if user is valid
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
