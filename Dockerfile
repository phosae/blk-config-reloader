FROM golang:1.18.4-alpine3.15 as builder

WORKDIR /go/src/github.com/phosae/blk-config-reloader/

COPY main.go .
COPY go.mod .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o blk-config-reloader .


# sudo docker run --privileged -v /run/systemd/system:/run/systemd/system -v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket -it ubuntu:18.04 systemctl
FROM ubuntu:18.04

RUN apt-get update -y
RUN apt-get install -y python-jsonpatch
RUN wget https://github.com/sclevine/yj/releases/download/v5.1.0/yj-linux-amd64 -O yj
RUN chmod +x yj
RUN mv yj /usr/bin/yj

COPY --from=builder /go/src/github.com/phosae/blk-config-reloader/ .

CMD ["./blk-config-reloader"]
