FROM golang:1.22.3

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /application

EXPOSE 8081

CMD ["/application"]