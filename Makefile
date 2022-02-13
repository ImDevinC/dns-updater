build:
	GOOS=linux GOARCH=arm go build -o updater ./cmd/updater/main.go

release: build
	tar czvf updater-${RELEASE_VERSION}.tar.gz updater

clean:
	rm -rf updater
	rm -rf updater*.tar.gz