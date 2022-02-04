FROM golang:alpine
ENV GOOS linux
ENV GOARCH arm
ENV GOARM 6
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -ldflags "-s" .
