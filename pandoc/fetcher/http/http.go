package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/go-pandoc/pandoc/fetcher"
)

type HttpFetcher struct {
	client *http.Client
}

type Params struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Data    []byte            `json:"data"`
	Replace map[string]string `json:"replace"`
}

func (p *Params) Validation() (err error) {
	if len(p.URL) == 0 {
		err = fmt.Errorf("[fetcher-http]: params of url is empty")
		return
	}

	p.Method = strings.ToUpper(p.Method)

	if len(p.Method) == 0 {
		p.Method = "GET"
	}

	if p.Method != "GET" && p.Method != "POST" {
		err = fmt.Errorf("[fetcher-http]: method %s not support", p.Method)
		return
	}

	return
}

func init() {
	err := fetcher.RegisterFetcher("http", NewHttpFetcher)

	if err != nil {
		panic(err)
	}
}

func NewHttpFetcher(conf config.Configuration) (httpFetcher fetcher.Fetcher, err error) {

	// Example: Adding Let's Encrypt certificate to the trust store
	certPool, err := x509.SystemCertPool()
	if err != nil {
		// Handle error
	}

	// Read the certificate file
	caCert, err := ioutil.ReadFile("path/to/letsencrypt_authority_x3.pem")
	if err != nil {
		// Handle error
	}

	// Append the certificate to the cert pool
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		// Handle error
	}

	// Use the certPool in your TLS configuration
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}
	httpFetcher = &HttpFetcher{
		client: httpClient,
	}
	return
}

func (p *HttpFetcher) Fetch(fetchParams fetcher.FetchParams) (data []byte, err error) {

	params := Params{}

	err = fetchParams.Unmarshal(&params)
	if err != nil {
		return
	}

	err = params.Validation()
	if err != nil {
		return
	}

	data, err = p.send(params)

	return
}

func (p *HttpFetcher) send(params Params) (data []byte, err error) {

	body := bytes.NewBuffer(params.Data)

	req, err := http.NewRequest(params.Method, params.URL, body)

	if err != nil {
		return
	}

	if len(params.Headers) > 0 {
		for k, v := range params.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := p.client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("[fetcher-http]: fetch url by %s failure <%s>, status code is %d", params.Method, params.URL, resp.StatusCode)
		return
	}

	data, err = ioutil.ReadAll(resp.Body)

	for k, v := range params.Replace {
		data = bytes.Replace(data, []byte(k), []byte(v), -1)
	}

	if err != nil {
		return
	}

	return
}
