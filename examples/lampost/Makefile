default: image

server: server/server.go
	# Building for Linux for the Docker container
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' server/ -o server/server

image: server/server server/Dockerfile
	cd server; docker build -t rgbkrk/iolohub .

GOARCH=amd64

windows : GOOS = windows
windows : STATIC_PATH = server/static/windows/amd64/lamposter.exe

linux : GOOS = linux
linux : STATIC_PATH = server/static/linux/amd64/lamposter

darwin : GOOS = darwin
darwin : STATIC_PATH = server/static/darwin/amd64/lamposter

darwin: $(STATIC_PATH)
	cd client; CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -tags netgo -ldflags '-w' -o ../$(STATIC_PATH) .

linux: $(STATIC_PATH)
	cd client; CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -tags netgo -ldflags '-w' -o ../$(STATIC_PATH) .

windows: $(STATIC_PATH)
	cd client; CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -tags netgo -ldflags '-w' -o ../$(STATIC_PATH) .

clients: darwin windows linux

upload: image
	docker push rgbkrk/iolohub
