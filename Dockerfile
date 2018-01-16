FROM scratch

COPY dist/aerokubot /usr/bin/

ENTRYPOINT ["/usr/bin/aerokubot"]
