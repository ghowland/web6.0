SHELL := /bin/bash
.DEFAULT_GOAL:=watch-run

GO_BUILD_VERSION="1.9.2"
BUILD_ARCH=amd64
NAME=web6
VERSION=0.0.1

WATCH_FILES=find . -type d \( -path \./vendor \) -prune -o -print | \
	grep -i .go 2> /dev/null

.PHONY: run
run:
	go run cmd/web6/web6.go

.PHONY: clean
clean:
	-rm -f bin/web6
	-rm -f web6_$(VERSION)_$(BUILD_ARCH).deb

.PHONY: check_go_version
check_go_version:
	@{ \
	goversion=$$(go version | cut -d' ' -f3) ; \
	if [ "$$goversion" != "go$(GO_BUILD_VERSION)" ]; then \
	    echo "Please build with golang $(GO_BUILD_VERSION)" ; \
	    exit 1 ; \
	fi \
	}

.PHONY: build
build: check_go_version clean
	GOOS=linux GOARCH=$(BUILD_ARCH) go build -o bin/web6 cmd/web6/web6.go

.PHONY: package
package: build
	fpm --input-type dir --output-type deb \
		--name $(NAME) \
		--version $(VERSION) \
		--license MIT \
		--maintainer shaw@vranix.com \
		--deb-group web6 \
		--before-install dist/before-install.sh \
		--after-remove dist/after-remove.sh \
		bin/web6=/usr/bin/web6 \
		dist/web6.template.json=/etc/web6/web6.template.json \
		dist/web6-upstart.conf=/etc/init/web6.conf

.PHONY: entr-warn
entr-warn:
	@echo "-------------------------------------------------"
	@echo " ! File watching functionality non-operational ! "
	@echo "                                                 "
	@echo " Install entr(1) to run tasks on file change.    "
	@echo " See http://entrproject.org/                     "
	@echo "-------------------------------------------------"

.PHONY: watch-run
watch-run:
	if command -v entr > /dev/null; then ${WATCH_FILES} | \
	entr -rc $(MAKE) run; else $(MAKE) run entr-warn; fi
