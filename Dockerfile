FROM golang:1.18.4-alpine3.15 as builder

WORKDIR /go/src/github.com/phosae/blk-config-reloader/

COPY main.go .
COPY go.mod .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o blk-config-reloader .

FROM ubuntu:18.04

RUN apt-get update -y
RUN apt-get install -y python-jsonpatch systemd wget
RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in ; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done);
RUN rm -rf /lib/systemd/system/multi-user.target.wants/;
RUN rm -rf /etc/systemd/system/.wants/;
RUN rm -rf /lib/systemd/system/local-fs.target.wants/;
RUN rm -rf /lib/systemd/system/sockets.target.wants/udev;
RUN rm -rf /lib/systemd/system/sockets.target.wants/initctl;
RUN rm -rf /lib/systemd/system/basic.target.wants/;
RUN rm -rf /lib/systemd/system/anaconda.target.wants/*;
VOLUME [ "/sys/fs/cgroup" ]

RUN wget https://github.com/sclevine/yj/releases/download/v5.1.0/yj-linux-amd64 -O yj
RUN chmod +x yj
RUN mv yj /usr/bin/yj

COPY --from=builder /go/src/github.com/phosae/blk-config-reloader/ .

CMD ["./blk-config-reloader"]
