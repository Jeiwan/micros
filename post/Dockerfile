FROM golang:1.10-alpine as builder
RUN apk add --update make protobuf git && \
    go get -u github.com/golang/protobuf/protoc-gen-go
WORKDIR /go/src/github.com/Jeiwan/micros/post
COPY . .
RUN make build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o micro_post .

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/Jeiwan/micros/post/micro_post /usr/local/bin
CMD ["micro_post"]
