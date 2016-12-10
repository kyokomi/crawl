package crawlhtml

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var charsetTransformMap = map[string]transform.Transformer{
	"EUC-JP":      japanese.EUCJP.NewDecoder(),
	"ISO-2022-JP": japanese.ISO2022JP.NewDecoder(),
	"Shift_JIS":   japanese.ShiftJIS.NewDecoder(),
}

// Crawler HTML scraping crawler
type Crawler struct {
	httpClient *http.Client
	headers    map[string]string
}

// New return Crawler
func New(roundTripper http.RoundTripper) *Crawler {
	return &Crawler{
		httpClient: &http.Client{Transport: roundTripper},
		headers:    map[string]string{},
	}
}

// SetHeader scraping request header
func (c *Crawler) SetHeader(key, value string) {
	c.headers[key] = value
}

// HTML scraping html
func (c *Crawler) HTML(uri string) (io.Reader, error) {
	body, err := c.crawlHTML(uri)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return readHTMLWithTransform(body)
}

func (c *Crawler) crawlHTML(linkURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func readHTMLWithTransform(r io.ReadCloser) (io.Reader, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// TODO: support japanese only
	result, err := chardet.NewHtmlDetector().DetectBest(data)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)

	// don't transform UTF-8
	if result.Charset == "UTF-8" {
		return reader, nil
	}

	return transformJapaneseDecode(result.Charset, reader)
}

func transformJapaneseDecode(charset string, reader io.Reader) (io.Reader, error) {
	t, ok := charsetTransformMap[charset]
	if !ok {
		return nil, fmt.Errorf("not supported charset = [%s]", charset)
	}
	return transform.NewReader(reader, t), nil
}
