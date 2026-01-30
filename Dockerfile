FROM ghcr.io/brandonkowalski/quasimodo:latest

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .

RUN GOWORK=off go build -v \
    -tags nodefaultfont \
    -o logjack ./app

CMD ["/bin/bash"]
