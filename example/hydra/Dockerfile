FROM golang:1.6

RUN go get github.com/Masterminds/glide
RUN go get github.com/ory-am/hydra

WORKDIR /go/src/github.com/ory-am/hydra

RUN glide install
RUN go install github.com/ory-am/hydra

ENTRYPOINT sleep 3; /go/bin/hydra host --dangerous-auto-logon

EXPOSE 4444
