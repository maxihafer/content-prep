FROM gcr.io/distroless/static-debian12
ENTRYPOINT ["/content-prep"]
COPY content-prep /