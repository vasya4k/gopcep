FROM alpine:3.14
COPY gopcep /.
COPY gopcep.toml /.
ENTRYPOINT ["./gopcep"]