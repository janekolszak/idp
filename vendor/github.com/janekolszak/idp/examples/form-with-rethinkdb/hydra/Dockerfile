FROM golang:1.6

RUN go get github.com/Masterminds/glide
RUN go get -d github.com/ory-am/hydra

RUN mkdir -p /etc/hydra
ADD key.pem /etc/hydra
ADD cert.pem /etc/hydra

WORKDIR /go/src/github.com/ory-am/hydra
RUN glide install
RUN go install .

ENTRYPOINT /go/bin/hydra host --dangerous-auto-logon
EXPOSE 4444
