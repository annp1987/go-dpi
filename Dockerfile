FROM golang:1.14
RUN apt-get update
RUN apt-get -y  install software-properties-common wget
#RUN add-apt-repository unive
RUN wget http://apt-stable.ntop.org/20.04/all/apt-ntop-stable.deb
RUN apt-get -y install ./apt-ntop-stable.deb
RUN apt-get update
RUN apt-get install -y pfring pfring-dkms build-essential autoconf automake libtool git libpcap-dev liblinear3 liblinear-dev libjson-c-dev
RUN git clone --branch 3.2-stable https://github.com/ntop/nDPI/ /tmp/nDPI
RUN cd /tmp/nDPI && ./autogen.sh && ./configure && make && make install && cd -

RUN mkdir -p $GOPATH/github.com/annp1987/go-dpi
WORKDIR $GOPATH/src/github.com/annp1987/go-dpi
ADD . .
RUN go build ./... && \
    go test ./... && \
    go test -bench=. && \
    go install ./godpi_example

ENTRYPOINT ["godpi_example"]
