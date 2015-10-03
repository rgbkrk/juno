# This Dockerfile is solely for building a Linux version of the client binding
FROM golang

RUN apt-get update
RUN apt-get install -y pkg-config
#RUN apt-get install -y libzmq3 libzmq3-dev pkg-config libsodium-dev libsodium13

RUN wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.3.tar.gz
RUN wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.3.tar.gz.sig
RUN wget https://download.libsodium.org/jedi.gpg.asc
RUN gpg --import jedi.gpg.asc
RUN gpg --verify libsodium-1.0.3.tar.gz.sig libsodium-1.0.3.tar.gz
RUN tar zxvf libsodium-1.0.3.tar.gz
RUN cd libsodium-1.0.3 && ./configure && make check && make install && ldconfig

RUN wget http://download.zeromq.org/zeromq-4.1.1.tar.gz
RUN tar zxvf zeromq-4.1.1.tar.gz
RUN cd zeromq-4.1.1 && ./configure --with-libsodium; make; make check && make install && ldconfig

RUN wget http://download.zeromq.org/czmq-3.0.1.tar.gz
RUN tar zxvf czmq-3.0.1.tar.gz
RUN cd czmq-3.0.1 && ./configure; make check && make install && ldconfig

RUN go get github.com/zeromq/goczmq

RUN go get github.com/rgbkrk/juno

ADD client.go $GOPATH/src/lamposter/client.go

WORKDIR $GOPATH/src/lamposter/ 

#RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' . -o client

# We `cat` the client binary so it can be used from the outside
# like docker run juno/lamposter > linux/amd64/lamposter
CMD ["cat", "/lamposter/client"]
