REPO := github.com/launchtime/scrapemonster

all:
	go install $(REPO)/cmd/crawl
	go install $(REPO)/cmd/dumpSnapshots
	go install $(REPO)/cmd/getDealInfo

deps:
	go get code.google.com/p/go.net/html
	go get code.google.com/p/cascadia
	go get github.com/ziutek/mymysql/godrv
