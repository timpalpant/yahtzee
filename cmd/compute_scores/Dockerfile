FROM golang:1.10

WORKDIR /go/src/github.com/timpalpant/yahtzee
COPY . .

RUN go get -d -v ./cmd/compute_scores/...
RUN go install -v ./cmd/compute_scores

CMD ["/go/bin/compute_scores", "-logtostderr", "-v", "1"]
