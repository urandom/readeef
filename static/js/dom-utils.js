(function() {
    "use strict";

    if (!Element.prototype.matches) {
        Element.prototype.matches = function(selector) {
            if (this.webkitMatchesSelector) {
                return this.webkitMatchesSelector(selector);
            } else if (this.mozMatchesSelector) {
                return this.mozMatchesSelector(selector);
            } else if (this.msMatchesSelector) {
                return this.msMatchesSelector(selector);
            } else if (this.oMatchesSelector) {
                return this.oMatchesSelector(selector);
            }
            return false;
        }
    }

    Element.prototype.closest = function(selector) {
        var element = this;

        do {
            if (element.matches(selector)) {
                return element;
            }
        } while (element = element.parentElement);
    }
})();
