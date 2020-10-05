FROM golang

EXPOSE 8080

VOLUME [ "./db/", "./conf/" ]

WORKDIR "linkmc/"
# download all the modules

RUN go get go.etcd.io/bbolt/...

COPY . .

RUN go mod vendor

RUN go install .

ENTRYPOINT [ "linkmc" ]