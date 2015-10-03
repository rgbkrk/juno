# This Dockerfile is solely for building a Linux version of the client binding
FROM golang

RUN apt-get install -y zmq-dev

ADD client.go /lamposter/client.go
WORKDIR /lamposter
RUN go get -u .
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' . -o client

# We `cat` the client binary so it can be used from the outside
# like docker run juno/lamposter > linux/amd64/lamposter
CMD ["cat", "/lamposter/client"]
