FROM alpine:latest

ADD {{project}}.tar.xz /
WORKDIR /{{project}}
CMD ["./{{project}}"]
