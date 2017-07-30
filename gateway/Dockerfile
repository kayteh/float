FROM alpine:latest
RUN apk add --no-cache su-exec ca-certificates &&\
    adduser -S float

COPY ./gateway /usr/bin/gateway
CMD ["su-exec", "float", "gateway"]