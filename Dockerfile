FROM golang:1.16-alpine

ENV GO111MODULE=on
ENV PORT=3000
ENV GOPROXY=https://goproxy.io
WORKDIR /app/server
# ENV GOPATH=/app/server
COPY go.mod .
COPY go.sum .

# RUN wget google.com
RUN ls 
RUN pwd
RUN go mod download
COPY . .

RUN go build -o main .
CMD ["./main"]