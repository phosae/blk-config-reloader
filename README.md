## containerd blkio config reloader

1. reload blkio.yaml from /<src>/default or /<src>/<node> to /etc/containerd/blkio.yaml if file md5sum changed 
2. make sure /etc/containerd/config.toml enables blkio control
3. systemctl restart containerd


## systemd

```
docker run --rm --privileged \
-v /run/systemd/system:/run/systemd/system \
-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
-v /etc/containerd/:/etc/containerd/ \
-v /your/blk/configs/:/configs/ \
-it zengxu/blk-config-reloader bash
``` 
