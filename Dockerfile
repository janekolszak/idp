FROM golang:1.6

ADD . /go/src/github.com/janekolszak/idp
WORKDIR /go/src/github.com/janekolszak/idp

RUN go get github.com/Masterminds/glide
RUN glide install
RUN go install github.com/janekolszak/idp

ENTRYPOINT /go/bin/idp -conf /root/.hydra.yml

EXPOSE 4444