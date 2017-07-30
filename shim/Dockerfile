FROM alpine:latest
RUN apk add --no-cache su-exec ca-certificates &&\
    adduser -S float

COPY ./shim /usr/bin/shim
CMD ["su-exec", "float", "shim"]