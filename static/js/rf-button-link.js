(function() {
    Polymer('rf-button-link', {
        href: '',
        target: '',

        upAction: function(event) {
            this.super(event);
            if (!event.defaultPrevented) {
                this.$.link.click();
            }
        },

        openInBackground: function() {
            if (navigator.userAgent.toLowerCase().indexOf('webkit') > -1) {
                var event = document.createEvent("MouseEvents");

                // Mouse click with ctrl key opens in background
                event.initMouseEvent("click", true, true, window, 0, 0, 0, 0, 0,
                    true, false, false, true, 0, null);

                this.$.link.dispatchEvent(event);
            } else {
                this.$.link.click();

                if (this.target) {
                    var target = window.open('', this.target);
                    if (target)  {
                        target.blur();
                    }
                }
                window.focus();
            }
        }
    });
})();
