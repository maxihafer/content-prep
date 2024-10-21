FROM alpine

ENTRYPOINT ["/usr/bin/content-prep"]

COPY content-prep /usr/bin/content-prep