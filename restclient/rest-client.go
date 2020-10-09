package restclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Body string
	Code int
}

func request(method string, serviceAddr string, apiPath string, timeout int64, payload interface{}, kwarg map[string]string) (Response, error) {
	var protocol string
	protocol, ok := kwarg["protocol"]
	if !ok {
		protocol = "http"
	}
	destURL := buildURL(protocol, serviceAddr, apiPath)
	proxyURL := getEnvProxy()
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Duration(timeout)*time.Second) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)) //设置发送接收数据超时
				return c, nil
			},
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal payload occur error: %s\n", err)
		return Response{}, err
	}
	body := bytes.NewReader(data)
	req, err := http.NewRequest(strings.ToUpper(method), destURL, body)
	if err != nil {
		return Response{}, err
	}
	token, ok := kwarg["token"]
	if ok {
		req.Header.Add("Authorization", "token "+token)
	}
	contentType, ok := kwarg["Content-Type"]
	if ok {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	accept, ok := kwarg["Accept"]
	if ok {
		req.Header.Set("Accept", accept)
	} else {
		req.Header.Set("Accept", "application/json")
	}
	// todo: add opentracing

	response, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}
	// fmt.Println(string(resBody))
	return Response{string(resBody), response.StatusCode}, nil
}

func parseTimeout(kwargs ...map[string]string) (map[string]string, int64) {
	kwarg := make(map[string]string)
	if len(kwargs) > 0 {
		kwarg = kwargs[0]
	}
	var timeout int64
	timeout = 10
	timeoutStr, ok := kwarg["timeout"]
	if ok {
		timeout, _ = strconv.ParseInt(timeoutStr, 10, 64)
	}
	return kwarg, timeout
}

func getEnvProxy() *url.URL {
	var httpProxy string
	httpProxy = os.Getenv("http_proxy")
	if httpProxy == "" {
		httpProxy = os.Getenv("HTTP_PROXY")
	}
	if httpProxy != "" {
		proxyURL, err := url.Parse(httpProxy)
		if err != nil {
			log.Printf("parse http proxy: %s occur error: %s\n", httpProxy, err)
			return nil
		}
		return proxyURL
	}
	return nil
}

func Get(serviceAddr string, apiPath string, payload interface{}, kwargs ...map[string]string) (Response, error) {
	kwarg, timeout := parseTimeout(kwargs...)
	return request("GET", serviceAddr, apiPath, timeout, payload, kwarg)
}

func Post(serviceAddr string, apiPath string, payload interface{}, kwargs ...map[string]string) (Response, error) {
	kwarg, timeout := parseTimeout(kwargs...)
	return request("POST", serviceAddr, apiPath, timeout, payload, kwarg)
}

func Put(serviceAddr string, apiPath string, payload interface{}, kwargs ...map[string]string) (Response, error) {
	kwarg, timeout := parseTimeout(kwargs...)
	return request("PUT", serviceAddr, apiPath, timeout, payload, kwarg)
}

func Patch(serviceAddr string, apiPath string, payload interface{}, kwargs ...map[string]string) (Response, error) {
	kwarg, timeout := parseTimeout(kwargs...)
	return request("PATCH", serviceAddr, apiPath, timeout, payload, kwarg)
}

func Delete(serviceAddr string, apiPath string, payload interface{}, kwargs ...map[string]string) (Response, error) {
	kwarg, timeout := parseTimeout(kwargs...)
	return request("DELETE", serviceAddr, apiPath, timeout, payload, kwarg)
}