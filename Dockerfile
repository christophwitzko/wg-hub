FROM scratch

COPY "./wg-hub" /
ENTRYPOINT ["/wg-hub"]
