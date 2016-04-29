(function(root) {
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
			},
			topLevelNavigation: {
				type: String,
				observer: '_topLevelNavigationChanged',
			},
			loginRedirect: {
				type: String,
			},
        },
		behaviors: [
			RouteBehavior,
			ThemeManagerBehavior,
			ConnectionUserBehavior,
		],
		_routingStarted: false,
        _state: 0,

        attached: function() {
			Excess.RouteManager.start();
			this._routingStarted = true;

            this.async(function() {
                if (!this.user && (this._state & state.VALIDATING) != state.VALIDATING) {
                    if (this.topLevelNavigation == "splash" || !location.pathname) {
						Excess.RouteManager.transitionTo('@login');
                    } else if (this.topLevelNavigation != "login") {
						Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
                    }
                }
            });

            document.addEventListener('rf-lazy-insert', function(event) {
                Polymer.updateStyles();
            }.bind(this));
        },

        onUserLoad: function(event, detail) {
            var storage = event.target;

            if (storage.value) {
                if (!storage.value.authTime || new Date().getTime() - storage.value.authTime > this.userTTL) {
                    storage.value = null;
                }
            }

            if (!detail.externalChange) {
                this.validateUser(storage.value);
            }
        },

        validateUser: function(user) {
            if (!user) {
                return;
            }

            this._state |= state.VALIDATING;

            var authCheck = this.$['auth-check'];

			if (!authCheck.connection) {
				setTimeout(function() {
					this.validateUser(user);
				}.bind(this), 100);
				return;
			}

			this.getConnection().user = user;
            authCheck.send();
        },

		validateRedirect: function() {
			if (!this._routingStarted) {
				this.debounce('validate-redirect', this.validateRedirect, 50);
				return;
			}

			if (this.topLevelNavigation == "login" && this.loginRedirect) {
				try {
					var redirect = this.decodeURI(this.loginRedirect),
						lang = this.user.ProfileData.language;

					if (lang && (redirect.substr(0, 2 + lang.length) == "/" + lang + "/")) {
						redirect = redirect.substr(1 + lang.length);
					}
					Excess.RouteManager.transitionTo(redirect);
				} catch(e) {
					Excess.RouteManager.transitionTo('@feed-all');
				}
			} else if (this.topLevelNavigation == "login") {
				Excess.RouteManager.transitionTo('@feed-all');
			} else if (this.topLevelNavigation != "feed" &&
					this.topLevelNavigation != "settings") {
				Excess.RouteManager.transitionTo('@feed-all');
			}
		},

        logout: function() {
            this.$.logout.send();
            this.$['auth-check'].close();
            this._setUser(null);
            this.async(function() {
				location.href = Excess.RouteManager.getRoutePath('@login');
            }, 250);
        },

        connectionUnauthorized: function() {
            this._state &= ~state.VALIDATING;
			switch (this.topLevelNavigation) {
			case "feed":
			case "settings":
				Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
				break;
			case "login":
                var login = Polymer.dom(this.root).querySelector('rf-login');
                if (login) {
                    login.invalid = true;
                }
				break;
			default:
				Excess.RouteManager.transitionTo('@login');
				break;
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
        },

		onAuthCheckMessage: function(event) {
			if (!event.detail.arguments.Auth) {
				return this.connectionUnauthorized();
			}

			var requestUser = this.getConnection().user;

			var user = event.detail.arguments.User;

			user.MD5API = requestUser.MD5API;
			user.authTime = new Date().getTime();
			user.capabilities = event.detail.arguments.Capabilities;

			user.ProfileData = user.ProfileData || {};

			if (('language' in user.ProfileData) && user.ProfileData.language != this.dataset.language) {
				location.href = location.href.replace('/' + this.dataset.language + '/', '/' + user.ProfileData.language + '/');
			}

			this._setUser(user);
			this._state &= ~state.VALIDATING;

			if (user.ProfileData.theme) {
				this.applyTheme();
			}

			if (user.ProfileData.shareServices) {
				user.ProfileData.shareServices.forEach(function(name) {
					RfShareServices.get(name).active = true;
				});
			}

			this.debounce('validate-redirect', this.validateRedirect);
		},

		_topLevelNavigationChanged: function(value, old) {
			if (value == "logout") {
				this.logout();
                return;
			}

            if (!this.user && (this._state & state.VALIDATING) != state.VALIDATING && value != "login") {
				if (this.topLevelNavigation == "splash" || !location.pathname) {
					Excess.RouteManager.transitionTo('@login');
				} else if (this.topLevelNavigation != "login") {
					Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
				}
			}
		},
    });
})(window);
