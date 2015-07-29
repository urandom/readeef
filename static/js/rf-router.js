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
                readOnly: true,
                notify: true
            }
        },
        _state: 0,

        attached: function() {
            this.async(function() {
                if (!this.user && (this._state & state.VALIDATING) != state.VALIDATING) {
                    if (!MoreRouting.getRouteByName('splash').children[0].active) {
                        MoreRouting.navigateTo('login');
                    } else if (!MoreRouting.getRouteByName('login').active &&
                        !MoreRouting.getRouteByName('login-from').active) {
                        MoreRouting.navigateTo('login-from', {url: this.encodeURI(location.pathname)});
                    }
                }
            });

            document.addEventListener('rf-lazy-insert', function(event) {
                if (event.detail.element.nodeName.toLowerCase() == 'rf-settings-base') {
                    event.detail.element.user = this.user;
                }
                Polymer.updateStyles();
            }.bind(this));
        },

        onRouteChange: function(event, detail) {
            if (!this.user & (this._state & state.VALIDATING) != state.VALIDATING) {
                if (MoreRouting.getRouteByName('login').active) {
                    this.$.splash.selected = 0;
                } else if (!MoreRouting.getRouteByName('splash').children[0].active) {
                    MoreRouting.navigateTo('login');
                } else if (!MoreRouting.getRouteByName('login').active &&
                    !MoreRouting.getRouteByName('login-from').active) {
                    MoreRouting.navigateTo('login-from', {url: this.encodeURI(location.pathname)});
                }
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
            if (!user) {
                return;
            }

            this._state |= state.VALIDATING;

            var authCheck = this.$['auth-check'];
            var validateMessage = function(event) {
                if (!event.detail.arguments.Auth) {
                    return this.connectionUnauthorized();
                }

                var user = event.detail.arguments.User;
                user.authTime = new Date().getTime();

                this._setUser(user);
                this._state &= ~state.VALIDATING;

                if (MoreRouting.getRouteByName('login-from').active) {
                    var login = Polymer.dom(this.root).querySelector('rf-login');
                    if (login) {
                        login.hide();
                    }

                    var url = MoreRouting.getRouteByName('login-from').params.url;

                    try {
                        MoreRouting.navigateTo(this.decodeURI(url));
                    } catch(e) {
                        MoreRouting.navigateTo('feed', {tagOrId: 'all'});
                    }
                } else if (MoreRouting.getRouteByName('login').active) {
                    var login = Polymer.dom(this.root).querySelector('rf-login');
                    if (login) {
                        login.hide();
                    }
                    MoreRouting.navigateTo('feed', {tagOrId: 'all'});
                } else if (MoreRouting.getRouteByName('splash').active) {
                    MoreRouting.navigateTo('feed', {tagOrId: 'all'});
                }
                this.$.splash.selected = 0;

                if (user.ProfileData.theme) {
                    document.body.classList.add('theme-' + user.ProfileData.theme);
                }
                authCheck.removeEventListener('rf-api-message', validateMessage);
            }.bind(this);

            authCheck.user = user;
            authCheck.addEventListener('rf-api-message', validateMessage);
            authCheck.send();
        },

        logout: function() {
            this._setUser(null);
            MoreRouting.navigateTo('login');
        },

        connectionUnauthorized: function() {
            if (!MoreRouting.getRouteByName('login').active) {
                MoreRouting.navigateTo('login-from', {url: location.pathname});
            } else {
                var login = Polymer.dom(this.root).querySelector('rf-login');
                if (login) {
                    login.invalid = true;
                }
            }
        },

        unhandledAPIError: function(data) {
            this.$['api-error'].text = "Error: " + JSON.stringify(data.error) + ", type: " + data.errorType;
            this.$['api-error'].show();
        },

        encodeURI: function(uri) {
            return encodeURIComponent(uri).replace(/%/g, '$');
        },

        decodeURI: function(encodedURI) {
            return decodeURIComponent(encodedURI.replace(/\$/g, '%'));
        }

    })
})();
