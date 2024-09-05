package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"requestforward/filter"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	log "github.com/kataras/golog"
)

var (
	ReverseClientList   []*http.Client
	ReverseProxyUrlList []string
	timeout                      = 8
	UrlOutputFile       *os.File = nil

	defaultTransport *http.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:       time.Duration(timeout) * time.Second,
			KeepAlive:     30 * time.Second,
			FallbackDelay: time.Duration(500) * time.Millisecond,
			DualStack:     true,
		}).DialContext,
		MaxIdleConnsPerHost:   10,
		MaxIdleConns:          100,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   0,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

func NoProxyHandler(w http.ResponseWriter, req *http.Request) {
	if req.Host == "" {
		log.Warn("[Proxy] : Cannot handle requests without Host header, e.g., HTTP 1.0")
		return
	}
}
func ReqFilter(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	log.Infof("[Proxy] : Processing  %s %s", req.Method, req.URL.String())
	if filter.AllowedHostNameReg != nil {
		//white list mode
		if !filter.IsAllowHostName(req.URL.Hostname()) {
			log.Debugf("[Proxy] Skipping push %s , because it isn't a white list hostname", req.URL.String())
			return req, nil
		}
	} else if filter.IsDisallowHostName(req.URL.Hostname()) {
		log.Debugf("[Proxy] Excluded hostname : %s , Skipping push url to server : %s", req.URL.Hostname(), req.URL.String())
		return req, nil
	}

	if filter.IsDisallowedReqPath(req.URL.Path) {
		log.Debugf("[Proxy] Skipping push url to server : %s , Excluded Request Path : %s", req.URL.String(), req.URL.Path)
		return req, nil
	}
	var reqBody []byte = nil
	if req.Method == "POST" || req.Method == "PUT" {
		if req.Body != nil {
			reqBody, _ = ioutil.ReadAll(req.Body)
			req.Body.Close()
			newBodyReader := bytes.NewReader(reqBody)
			req.ContentLength = int64(len(reqBody))
			req.Body = io.NopCloser(newBodyReader)
		}
	}
	var baseUrl string = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
	go func(reqBaseUrl string) {
		if UrlOutputFile != nil {
			UrlOutputFile.WriteString(reqBaseUrl + "\r\n")
		}
	}(baseUrl)

	go func(oriReq *http.Request, oriReqBody []byte) {
		log.Infof("[Proxy] : Pushing  %s  to %s", req.URL.String(), strings.Join(ReverseProxyUrlList, ","))

		for index, client := range ReverseClientList {
			newReq := oriReq.Clone(context.Background())
			//set header
			newReq.Header.Set("X-Forwarded-Host", oriReq.URL.Host)
			//清空RequestURI,避免client处理请求时报错
			newReq.RequestURI = ""

			//set body
			if newReq.Method == "POST" || newReq.Method == "PUT" && oriReqBody != nil {
				newBodyReader := bytes.NewReader(oriReqBody)
				newReq.ContentLength = int64(len(oriReqBody))
				newReq.Body = io.NopCloser(newBodyReader)
			}
			go func(proxyIndex int, proxyReq *http.Request, c *http.Client) {
				resp, err := c.Do(proxyReq)
				if err == nil {
					io.Copy(ioutil.Discard, resp.Body) //需要重用client,因此需要读取并丢弃body
					resp.Body.Close()                  //关闭body避免内存泄露
					log.Debugf("[Proxy] : %s : Proxy via  %s ,  HTTP %d", proxyReq.URL.String(), ReverseProxyUrlList[proxyIndex], resp.StatusCode)
				} else {
					log.Warnf("[Proxy] : %s : Proxy via  %s is failed ,%s ", proxyReq.URL.String(), ReverseProxyUrlList[proxyIndex], err)

				}

			}(index, newReq, client)

		}
		if oriReq.Body != nil {
			oriReq.Body.Close()
		}
	}(req.Clone(context.Background()), reqBody)
	return req, nil
}

func RespFilter(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp != nil {
		log.Infof("[Proxy] : Processed : %s : HTTP %d", ctx.Req.URL.String(), resp.StatusCode)
	}
	return resp
}
func New(pushProxyUrlList []string, urlOutFile *os.File, enableVerbose bool) *goproxy.ProxyHttpServer {
	if urlOutFile != nil {
		UrlOutputFile = urlOutFile
	}
	proxy := goproxy.NewProxyHttpServer()
	proxy.NonproxyHandler = http.HandlerFunc(NoProxyHandler)
	proxy.OnRequest().DoFunc(ReqFilter)
	proxy.OnResponse().DoFunc(RespFilter)
	proxy.OnRequest().HandleConnectFunc(goproxy.AlwaysMitm)
	proxy.Verbose = enableVerbose
	proxy.Tr = defaultTransport.Clone()
	SetProxyCA()

	if len(pushProxyUrlList) > 0 {
		ReverseClientList = make([]*http.Client, len(pushProxyUrlList))
		ReverseProxyUrlList = pushProxyUrlList
		for index, proxyUrl := range pushProxyUrlList {
			ReverseClientList[index] = &http.Client{
				Transport: defaultTransport.Clone(),
			}
			if len(proxyUrl) > 0 && !strings.HasPrefix(proxyUrl, "http") && !strings.HasPrefix(proxyUrl, "socks") {
				proxyUrl = "http://" + proxyUrl
			}
			urlObj, err := url.Parse(proxyUrl)
			if err != nil {
				log.Fatalf("[Proxy] Invalid proxy url :%s", proxyUrl)
			}
			proxyFunc := func(req *http.Request) (*url.URL, error) {
				return urlObj, nil
			}

			ReverseClientList[index].Transport.(*http.Transport).Proxy = proxyFunc
			log.Infof("[Proxy] Added reverse proxy %s", proxyUrl)
		}
	}
	return proxy
}
