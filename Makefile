format:
	gofumpt -l -w .

build:
	make format
	go build -ldflags '-w -s' -trimpath github.com/sqkam/sensitivecrawler

run:
	make build
	./sensitivecrawler
builddebug:
	make format
	go build -tags debug -o sensitivecrawlerdebug -ldflags '-w -s' -trimpath github.com/sqkam/sensitivecrawler