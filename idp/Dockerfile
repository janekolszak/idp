FROM golang:1.6

ADD . /go/src/github.com/janekolszak/hydra-idp-go
WORKDIR /go/src/github.com/janekolszak/hydra-idp-go

RUN go get github.com/Masterminds/glide
RUN glide install
RUN go install github.com/janekolszak/hydra-idp-go

ENTRYPOINT sleep 5 && /go/bin/hydra-idp-go

EXPOSE 4444