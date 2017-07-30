FROM alpine:latest
RUN apk add --no-cache su-exec ca-certificates &&\
    adduser -S float
    
COPY ./coordinator /usr/bin/coordinator
CMD ["su-exec", "float", "coordinator"]