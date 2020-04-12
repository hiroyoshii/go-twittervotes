FROM golang:1.12

RUN apt-get install git
RUN git clone https://github.com/hiroyoshii/go-twittervotes.git
WORKDIR /go/go-twittervotes
RUN go mod download
RUN go build ./twittervotes
