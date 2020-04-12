FROM golang:1.12

RUN apt-get install git
RUN git clone https://github.com/hiroyoshii/go-twittervotes.git
RUN cd go-twittervotes/twittervotes
RUN go build
RUN go run ./
