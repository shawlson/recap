PSQL               = /usr/local/pgsql/bin/psql
MKDIR              = /bin/mkdir -p

RECAP_DB          ?= recap
RECAP_DIR         ?= $(HOME)/recap
DIST               = dist

SRC_FILES          = recap.go \
                     internal/cli.go \
                     internal/writeHTML.go \
                     internal/html.go \
                     internal/models.go \
                     internal/store.go
DIST_BIN_DIR       = $(DIST)/bin
BINARIES           = $(DIST_BIN_DIR)/recap

SRC_TEMPLATES_DIR  = web/templates
DIST_TEMPLATES_DIR = $(DIST)/templates
TEMPLATES          = $(DIST_TEMPLATES_DIR)/header.tmpl \
                     $(DIST_TEMPLATES_DIR)/sidebar.tmpl \
                     $(DIST_TEMPLATES_DIR)/index.tmpl \
                     $(DIST_TEMPLATES_DIR)/game.tmpl

SRC_STATIC_DIR     = web/static
DIST_STATIC_DIR    = $(DIST)/www/static
STATIC_ASSETS      = $(DIST_STATIC_DIR)/recap.css
STATIC_ASSETS_GZ   = $(DIST_STATIC_DIR)/recap.css.gz

.PHONY: all deps db install clean

all: $(BINARIES) $(TEMPLATES) $(STATIC_ASSETS_GZ)
deps:
	go get
	go get ./minifier
db:
	$(PSQL) -v ON_ERROR_STOP=1 -f sql/database.sql $(RECAP_DB) && \
	$(PSQL) -v ON_ERROR_STOP=1 -f sql/seed.sql $(RECAP_DB)
install: all
	cp -R $(DIST)/. $(RECAP_DIR)
clean:
	rm -rf $(DIST)

$(DIST_BIN_DIR)/recap: $(SRC_FILES)
	@$(MKDIR) $(DIST_BIN_DIR)
	go build -o $(DIST_BIN_DIR)/recap

$(DIST_TEMPLATES_DIR)/header.tmpl: $(SRC_TEMPLATES_DIR)/header.tmpl
	@$(MKDIR) $(DIST_TEMPLATES_DIR)
	<$(SRC_TEMPLATES_DIR)/header.tmpl go run ./minifier -type=html > $@

$(DIST_TEMPLATES_DIR)/sidebar.tmpl: $(SRC_TEMPLATES_DIR)/sidebar.tmpl
	@$(MKDIR) $(DIST_TEMPLATES_DIR)
	<$(SRC_TEMPLATES_DIR)/sidebar.tmpl go run ./minifier -type=html > $@

$(DIST_TEMPLATES_DIR)/index.tmpl: $(SRC_TEMPLATES_DIR)/index.tmpl
	@$(MKDIR) $(DIST_TEMPLATES_DIR)
	<$(SRC_TEMPLATES_DIR)/index.tmpl go run ./minifier -type=html > $@

$(DIST_TEMPLATES_DIR)/game.tmpl: $(SRC_TEMPLATES_DIR)/game.tmpl
	@$(MKDIR) $(DIST_TEMPLATES_DIR)
	<$(SRC_TEMPLATES_DIR)/game.tmpl go run ./minifier -type=html > $@

$(STATIC_ASSETS_GZ): $(STATIC_ASSETS)

$(DIST_STATIC_DIR)/recap.css: $(SRC_STATIC_DIR)/recap.css
	@$(MKDIR) $(DIST_STATIC_DIR)
	<$(SRC_STATIC_DIR)/recap.css go run ./minifier -type=css > $@

$(DIST_STATIC_DIR)/recap.css.gz: $(DIST_STATIC_DIR)/recap.css
	@$(MKDIR) $(DIST_STATIC_DIR)
	gzip -k -9 $(DIST_STATIC_DIR)/recap.css
