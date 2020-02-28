FROM alpine:3.10

ADD sbanken-client static templates ./

ENTRYPOINT ["./sbanken-client"]
