FROM golang:alpine
WORKDIR /app
ADD . /app
RUN go mod download
RUN go build -o api.exe cmd/api/api.go
EXPOSE 5678
CMD ["/app/api.exe", "5678"]