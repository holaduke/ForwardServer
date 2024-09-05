package main

import (
	"flag"
	log "github.com/kataras/golog"
	"net/http"
	"net/url"
	"os"
	"requestforward/filter"
	proxyserver "requestforward/proxy"
	"requestforward/utils"
	"runtime"
	"strings"
)

func main() {
	var listenAddr string
	var pushToUrlList string
	var logLevel string
	var upstreamProxy string
	var saveFileName string
	var limit uint64
	var allowedHostnames string
	var extraDisallowReqPath string
	flag.StringVar(&listenAddr, "listen", ":18080", "the proxy server listen addr")
	flag.StringVar(&pushToUrlList, "push-to-proxy", "", "the push proxy , split by comma")
	flag.StringVar(&logLevel, "log-level", "info", "the log level : debug , info, warn , error")
	flag.StringVar(&upstreamProxy, "upstream-proxy", "", "upstream proxy , support scheme :  http , https, socks5, for example : http://127.0.0.1:8080 , socks5://127.0.0.1:1080")
	flag.StringVar(&saveFileName, "url-out", "", "The file that save the proxies url")
	flag.Uint64Var(&limit,"max-open-files",10240,"the value of max open files of process (*unix only)")
	flag.StringVar(&allowedHostnames,"allow-hostname","","The allow hostname ,split by semicolon, for example : *test.com*;*example*;")
	flag.StringVar(&extraDisallowReqPath,"extra-disallow-path","","The extra disallowed request path keywords , split by comma")
	flag.Parse()

	curLimit := utils.SetFSLimit(limit)
	if curLimit == limit {
		if runtime.GOOS != "windows" {
			log.Infof("Set max open files of process to %d is successful",limit)
		}
	} else {
		log.Warnf("Set max open files of process to %d is failed",limit)
	}

	if len(allowedHostnames) != 0 {
		allowhostnameList := strings.Split(allowedHostnames,";")
		log.Infof("[Proxy] : The Whitelist mode is enabled, allowed hostname :%v",allowhostnameList)
		filter.DefaultAllowHostname  = allowhostnameList
	}
	if len(extraDisallowReqPath) > 0 {
		disallowReqPath := strings.Split(extraDisallowReqPath,",")
		log.Infof("[Proxy] : The URLs whose  request path contain these keywords will not be pushed to server  : %v ",disallowReqPath)
		filter.DefaultDisallowPath = append(filter.DefaultDisallowPath,disallowReqPath...)
	}

	filter.Init()
	log.SetLevel(logLevel)
	var urlOutFile *os.File = nil
	if len(saveFileName) != 0 {
		outFile, err := os.OpenFile(saveFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_SYNC, 0666)
		if err == nil {
			defer outFile.Close()
			log.Infof("[Proxy] Save proxied url to %s", saveFileName)
			urlOutFile = outFile
		} else {
			log.Fatalf("cannot create the file %s :%s",saveFileName,err)

		}
	}

	proxyUrlList := make([]string,0)
	if len(pushToUrlList ) > 0 {
		proxyUrlList = strings.Split(pushToUrlList,",")
	}
	var enableGoProxyVerbose = false
	if strings.ToLower(logLevel) == "debug" {
		log.Infof("[Proxy] Enable goproxy verbose output")
		enableGoProxyVerbose = true
	}
	proxy := proxyserver.New(proxyUrlList, urlOutFile,enableGoProxyVerbose)

	if len(upstreamProxy) > 0 {
		urlObj, err := url.Parse(upstreamProxy)
		if err == nil {
			log.Infof("[Proxy] Using upstream proxy : %s ...", upstreamProxy)
			proxyFunc := func(req *http.Request) (*url.URL, error) {
				return urlObj, nil
			}
			proxy.Tr.Proxy = proxyFunc
		} else {
			log.Warnf("[Proxy] Invalid upstream proxy : %s, ignoring...", upstreamProxy)
		}

	}

	log.Infof("[Proxy] : Start proxy at %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, proxy))
}
