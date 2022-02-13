build:
	go build -o updater ./cmd/updater/main.go

release: build
	tar czvf updater.tar.gz updater

clean:
	rm -rf updater
	rm -rf updater.tar.gz