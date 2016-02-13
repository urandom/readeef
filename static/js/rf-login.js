(function() {
    "use strict";

    Polymer({
        is: 'rf-login',
        behaviors: [
            Polymer.NeonAnimationRunnerBehavior,
			UserBehavior,
			ThemeBehavior,
        ],
        properties: {
            invalid: {
                type: Boolean,
                value: false
            },
            animationConfig: {
                value: function() {
                    return {
                        'entry': [{
                            name: 'slide-down-animation',
                            node: this.$['login-card']
                        }, {
                            name: 'fade-in-animation',
                            node: this.$['login-card']
                        }],

                        'exit': [{
                            name: 'slide-up-animation',
                            node: this.$['login-card']
                        }, {
                            name: 'fade-out-animation',
                            node: this.$['login-card']
                        }]
                    };
                }
            }
        },

        onKeypress: function(event) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == 'Enter' || code == 13) {
                if (event.target === this.$.login.$.input) { 
                    this.$.password.$.input.focus();
                } else if (event.target === this.$.password.$.input) {
                    this.$.submit.fire('tap');
                }
            }
        },

        onLogin: function() {
            var user = {
                Login: this.$.login.value,
                MD5API: CryptoJS.MD5(this.$.login.value + ":" + this.$.password.value).toString(CryptoJS.enc.Base64)
            };

            this.$.login.value = "";
            this.$.password.value = "";

            this.validateUser(user);
        }
    });
})();
