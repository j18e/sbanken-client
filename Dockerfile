FROM golang:1.13

ENV GOOS=linux
ENV GOARCH=386
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /bin/compiled

FROM alpine:3.10
COPY --from=0 /bin/compiled /bin/compiled
ENTRYPOINT ["/bin/compiled"]
