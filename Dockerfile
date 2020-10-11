FROM golang AS builder

RUN go get go.etcd.io/bbolt/...

COPY . .

ENV GOPATH "${WORKDIR}"

RUN go mod vendor

RUN go build

FROM alpine

RUN apk add --no-cache libc6-compat

# expose the port
EXPOSE 8080

VOLUME [ "/data" ]

WORKDIR "linkmc/"


COPY --from=builder "/go/linkmc" .
ENTRYPOINT [ "./linkmc" ]