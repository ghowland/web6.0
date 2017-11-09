.DEFAULT_GOAL:=watch-run

WATCH_FILES=find . -type d \( -path \./vendor \) -prune -o -print | \
	grep -i .go 2> /dev/null

run:
	go run cmd/web6/web6.go

clean:
	-rm -f bin/web6

build: clean
	GOOS=linux GOARCH=amd64 go build -o bin/web6 cmd/web6/web6.go

entr-warn:
	@echo "-------------------------------------------------"
	@echo " ! File watching functionality non-operational ! "
	@echo "                                                 "
	@echo " Install entr(1) to run tasks on file change.    "
	@echo " See http://entrproject.org/                     "
	@echo "-------------------------------------------------"


watch-run:
	if command -v entr > /dev/null; then ${WATCH_FILES} | \
	entr -rc $(MAKE) run; else $(MAKE) run entr-warn; fi
