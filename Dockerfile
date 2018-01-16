FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY dist/aerokubot /usr/bin/

ENTRYPOINT ["/usr/bin/aerokubot"]
