package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var configDir = flag.String("d", ".", "containerd config directory")
var srcDir = flag.String("s", ".", "source config directory")
var node = flag.String("s", "", "node name")

func main() {
	flag.Parse()

	if *node == "" {
		*node = os.Getenv("NODE_NAME")
	}
	stop := SetupSignalHandler()

	tik := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-tik.C:
			sync()
		case <-stop:
			os.Exit(0)
		}
	}
}

func sync() {
	blkfile := *srcDir + "/default"
	if *node != "" {
		blkfile = *srcDir + "/" + *node
	}

	if _, err := os.Stat(blkfile); err != nil {
		return // blk config file not exist
	}

	if checksum(blkfile) == checksum(*configDir+"/blkio.yaml") {
		return
	}
	fmt.Println("blkio config changed, start do sync...")

	err := exec.Command("cp", blkfile, *configDir+"/blkio.yaml").Run()
	if err != nil {
		panic(err)
	}

	ensureContainerdSrvConfig(*configDir)
	fmt.Println("restart containerd...")
	err = exec.Command("systemctl", "restart", "containerd").Run()
	if err == nil {
		fmt.Println("restart containerd ok")
	} else {
		fmt.Printf("err restart containerd:%v\n", err)
	}
}

type Config struct {
	Plugins map[string]map[string]interface{} `json:"plugins"`
}

func ensureContainerdSrvConfig(dir string) {
	f, err := os.Open(dir + "/config.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	jfname := dir + "/config.json.tmp"
	if _, err = os.Stat(jfname); err == nil {
		os.Remove(jfname)
	}
	jf, err := os.Create(jfname)
	if err != nil {
		panic(err)
	}

	var stderr bytes.Buffer
	yj := exec.Command("yj", "-tj")
	yj.Stdin = f
	yj.Stdout = jf
	yj.Stderr = &stderr
	err = yj.Run()
	fmt.Println(stderr.String())
	if err != nil {
		panic(err)
	}
	jf.Close()

	var cfg Config
	b, err := ioutil.ReadFile(jfname)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &cfg)

	if cfg.Plugins["io.containerd.service.v1.tasks-service"] == nil {
		os.Remove(dir + "/patch.json.tmp")
		var p = `
		[
			{
				"op": "add",
				"path": "/plugins/io.containerd.service.v1.tasks-service",
				"value": {"blockio_config_file": "/etc/containerd/blkio.yaml"}
			}
		]
		`
		err = ioutil.WriteFile(dir+"/patch.json.tmp", []byte(p), 0666)
		if err != nil {
			panic(err)
		}
		cmd := "jsonpatch config.json.tmp patch.json.tmp | yj -jt -i > config.toml.tmp && mv config.toml.tmp config.toml && rm *.tmp"
		fmt.Println(exec.Command("bash", "-c", cmd).Run())
	} else {
		fmt.Println("plugins.\"io.containerd.service.v1.tasks-service\" already specified, skip blkio.yaml path patch")
	}
}

func checksum(f string) string {
	file, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		panic(err)
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		panic(err)
	}

	return string(hash.Sum(nil))
}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}
