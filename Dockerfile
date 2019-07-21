FROM alpine

COPY carts /opt/carteria/carts
RUN ["apk", "add", "--no-cache", "ca-certificates"]

ENTRYPOINT /opt/carteria/carts
