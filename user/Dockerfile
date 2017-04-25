FROM golang:1.8

MAINTAINER Randy Cahyana "ranch@randycahyana.com"

RUN go get -u github.com/golang/dep/...

ADD . /go/src/github.com/rnd/kudu-service/user

WORKDIR /go/src/github.com/rnd/kudu-service/user

RUN dep ensure

RUN go install github.com/rnd/kudu-service/user

ENTRYPOINT /go/bin/user