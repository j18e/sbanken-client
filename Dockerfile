FROM alpine:3.10

ADD sbanken-client .

ENTRYPOINT ["./sbanken-client"]
