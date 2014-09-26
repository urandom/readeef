(function() {
    "use strict";

    Polymer('rf-article-display', {
        observe: {
          'articles template': 'initialize'
        },

        attached: function() {
            this.template = this.querySelector('template');
            this.template.setAttribute('id', 'article-template');

            this.templates = [];

            [0, 1, 2].forEach(function(index) {
                var template = document.createElement('template');
                template.setAttribute('ref', 'article-template');

                this.template.parentNode.insertBefore(template, this.template);
                this.templates[index] = template;
            }.bind(this));
        },

        initialize: function() {
            if (!this.articles || !this.template) {
                return;
            }

            this._originalArticleProperties = Object.getOwnPropertyNames(this.articles[0] || {});
            this._physicalArticles = [];
            for (var i = 0; i < 3; ++i) {
                this._physicalArticles[i] = {};
                this.updateItem(i, i);
            }
        },

        updateItem: function(virtualIndex, physicalIndex) {
            var src = this.articles[virtualIndex],
                dest = this._physicalArticles[physicalIndex],
                middle = 1;

            if (!src) {
                return;
            }

            dest.selected = physicalIndex == middle;
            dest.wide = this.wide;
            for (var i = 0, p; p = this._originalArticleProperties[i]; ++i) {
                dest[p] = src[p];
            }

            dest._physicalIndex = physicalIndex;
            dest._virtualIndex = virtualIndex;
        },

        articleChanged: function() {
            if (!this.article) {
                return;
            }

            var virtual = this.articles.indexOf(this.article),
                article, physical = -1, middle = 1;

            if (virtual == -1) {
                return;
            }

            for (var i = 0, a; a = this._physicalArticles[i]; ++i) {
                a.selected = false;
                if (a.Id == this.article.Id) {
                    a.selected = true;
                    physical = i;
                }
            }

            if (physical == 0) {
                this._physicalArticles.pop();
                this._physicalArticles.unshift({});
                this.updateItem(virtual - 1, 0);

                this.templates.unshift(this.templates.pop());
                this.templates[0].model = this._physicalArticles[0];
                this.cleanTemplateElement(this.templates[0]);
            } else if (physical == 2) {
                this._physicalArticles.shift();
                this._physicalArticles.push({});
                this.updateItem(virtual + 1, 2);

                this.templates.push(this.templates.shift());
                this.templates[2].model = this._physicalArticles[2];
                this.cleanTemplateElement(this.templates[2]);
            } else if (physical == -1) {
                for (var i = virtual - middle, j = 0; j < 3; ++i, ++j) {
                    this.updateItem(i, j);
                    this.templates[j].model = this._physicalArticles[j];
                    this.cleanTemplateElement(this.templates[j]);
                }
            }

            for (var i = 0, t; t = this.templates[i]; ++i) {
                if (!t._element) {
                    if (this.$.viewport.children[i]) {
                        this.$.viewport.insertBefore(t.createInstance(t.model), this.$.viewport.children[i]);
                    } else {
                        this.$.viewport.appendChild(t.createInstance(t.model));
                    }
                    t._element = this.$.viewport.children[i];
                    this.initializeElement(t._element, t.model);
                }

                t._element.style.display = i == middle ? '' : 'none';
            }
        },

        cleanTemplateElement: function(template) {
            if (template._element) {
                if (template._element.parentNode) {
                    template._element.parentNode.removeChild(template._element);
                }
                template._element = undefined;
            }
        },

        initializeElement: function(item, model) {
            var description = item.querySelector('.article-description'),
                imageStyler = function() {
                    if (image.width < 300) {
                        (image.parentNode || description).classList.add('clearfix');
                        image.style.float = "right";
                    }
                }, image;

            description.innerHTML = model.Description;
            image = description.querySelector('img');

            if (image) {
                if (image.complete) {
                    imageStyler();
                } else {
                    image.addEventListener('load', imageStyler);
                }
            }

            Array.prototype.forEach.call(
                description.querySelectorAll('img'),
                function(element) {
                    element.style.width = 'auto';
                    element.style.height = 'auto';
                }
            );
        },

        openCurrentArticle: function() {
            if (this.templates[1] && this.templates[1]._element) {
                this.templates[1]._element.querySelector('.article-link').openInBackground();
            }
        }
    });
})();
