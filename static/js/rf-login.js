(function() {
    "use strict";

    Polymer('rf-login', {
        invalid: false,

        onKeypress: function(event) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == 'Enter' || code == 13) {
                if (event.target === this.$.login) { 
                    this.$.password.focusAction();
                } else if (event.target === this.$.password) {
                    this.$.submit.asyncFire('tap');
                }
            }
        },

        onLogin: function() {
            this.user = {
                Login: this.$.login.value,
                MD5API: CryptoJS.MD5(this.$.login.value + ":" + this.$.password.value).toString(CryptoJS.enc.Base64)
            };

            this.$.login.value = "";
            this.$.password.value = "";
        }
    });
})();
