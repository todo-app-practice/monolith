FROM golang:1.24.3-alpine as build

WORKDIR /home/app

COPY . .

RUN go build -o main cmd/todo-app/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=build /home/app/main .

CMD ["./main"]