FROM golang:1.10-alpine

ARG SERVICE_NAME

ADD . /go/src/github.com/danielcuervo/kafka-101/${SERVICE_NAME}

WORKDIR /go/src/github.com/danielcuervo/kafka-101/${SERVICE_NAME}

RUN apk update \
    && apk upgrade \
    && apk add git \
    && apk add gcc \
    && go get -u github.com/golang/dep/cmd/dep \
    && dep init

CMD ["go", "run", "main.go"]