default: image

server: server.go
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

image: server Dockerfile
	docker build -t rgbkrk/iolohub .

upload: image
	docker push rgbkrk/iolohub
