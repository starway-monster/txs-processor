FROM golang:latest as build 

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o processor cmd/processor/main.go

FROM alpine:latest as production

WORKDIR /app

COPY --from=build /app/processor  /app/processor
CMD ["./processor"]