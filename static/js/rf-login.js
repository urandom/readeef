(function() {
    "use strict";

    Polymer({
        is: 'rf-login',
        properties: {
            invalid: {
                type: Boolean,
                value: false
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
            this.$.login.value = "";
            this.$.password.value = "";

            var user = {
                Login: this.$.login.value,
                MD5API: CryptoJS.MD5(this.$.login.value + ":" + this.$.password.value).toString(CryptoJS.enc.Base64)
            };

            Polymer.dom(document).querySelector('rf-router')._setUser(user);
        }
    });
})();
