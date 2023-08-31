FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY ./dist/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
