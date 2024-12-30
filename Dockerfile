FROM golang:1.23
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY assets/deleted_movie.jpg /app/assets/del_image.jpg
RUN go build -o main .
CMD ["./main"]