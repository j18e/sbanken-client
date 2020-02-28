FROM alpine:3.10

WORKDIR /work

ADD sbanken-client .
ADD static ./static
ADD templates ./templates

ENTRYPOINT ["./sbanken-client"]
