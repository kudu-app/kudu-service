FROM golang:1.8

MAINTAINER Randy Cahyana "ranch@randycahyana.com"

RUN go get -u github.com/golang/dep/...

ADD . /go/src/github.com/rnd/kudu-service/task

WORKDIR /go/src/github.com/rnd/kudu-service/task

RUN dep ensure

RUN go install github.com/rnd/kudu-service/task

ENTRYPOINT /go/bin/task