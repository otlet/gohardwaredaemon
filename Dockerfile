FROM golang:alpine

# Set go bin which doesn't appear to be set already.
ENV GOBIN /go/bin

# Install git
RUN apk add --no-cache git

# build directories
RUN mkdir /app
RUN mkdir /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app

# Go dep!
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

# Build my app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

ENV GHW_DISABLE_WARNINGS=1
CMD ["/app/main"]