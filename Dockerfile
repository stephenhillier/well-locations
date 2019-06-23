FROM golang:1.12
WORKDIR /go/src/github.com/stephenhillier/well-locations
ADD . /go/src/github.com/stephenhillier/well-locations
RUN go get ./...
RUN go install
EXPOSE 8000
CMD ["well-locations"]
