/*
 * @Author: gitsrc
 * @Date: 2020-11-09 16:31:33
 * @LastEditors: gitsrc
 * @LastEditTime: 2020-11-10 15:05:07
 * @FilePath: /influx-proxy-gitsrc/main.go
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/chengshiwen/influx-proxy/backend"
	"github.com/chengshiwen/influx-proxy/service"
	"github.com/chengshiwen/influx-proxy/util"
)

var (
	configFile string
	version    bool
	//GitCommit is commit id
	GitCommit = "not build"
	//BuildTime is build time
	BuildTime = "not build"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	log.SetOutput(os.Stdout)
	flag.StringVar(&configFile, "config", "proxy.json", "proxy config file")
	flag.BoolVar(&version, "version", false, "proxy version")
	flag.Parse()
}

func main() {
	if version {
		fmt.Printf("Version:    %s\n", backend.Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	cfg, err := backend.NewFileConfig(configFile) //Read config file and create config object.
	if err != nil {
		log.Printf("illegal config file: %s\n", err)
		return
	}
	log.Printf("version: %s, commit: %s, build: %s", backend.Version, GitCommit, BuildTime)
	cfg.PrintSummary()

	err = util.MakeDir(cfg.DataDir)
	if err != nil {
		log.Fatalln("create data dir error")
		return
	}

	//判断是够开启UDP-Server
	if cfg.UDPEnable {
		//开启UDP
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("UDP Server recover", r)
				}
			}()
			err := service.NewUDPService(cfg).ListenAndServe()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	mux := http.NewServeMux()
	service.NewHTTPService(cfg).Register(mux)

	server := &http.Server{
		Addr:        cfg.ListenAddr,
		Handler:     mux,
		IdleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
	}
	if cfg.HTTPSEnabled {
		log.Printf("https service start, listen on %s", server.Addr)
		err = server.ListenAndServeTLS(cfg.HTTPSCert, cfg.HTTPSKey)
	} else {
		log.Printf("http service start, listen on %s", server.Addr)
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Print(err)
		return
	}
}
