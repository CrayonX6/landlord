package acme

import (
    "bytes"
    logger "beginning/pkg/log"
    "encoding/json"
    "fmt"
    "github.com/ddliu/go-httpclient"
    "github.com/pkg/errors"
    "io"
    "io/ioutil"
)

type cURL struct {
    client   *httpclient.HttpClient
    response io.Reader
    code     int
}

func NewCURL() *cURL {
    return &cURL{client: httpclient.NewHttpClient().Begin()}
}

// cURL.自定义参数
func (c *cURL) Options(handler func(client *httpclient.HttpClient) *httpclient.HttpClient) *cURL {
    c.client = handler(c.client)
    return c
}

// cURL.Get
func (c *cURL) SendGet(url string, params map[string]string) (map[string]interface{}, error) {
    res, err := c.client.Get(url, params)
    if err != nil {
        return make(map[string]interface{}), err
    }

    c.response = res.Body
    c.code = res.StatusCode
    result, err := c.Parser()
    if err != nil {
        return make(map[string]interface{}), err
    }
    return result, nil
}

// cURL.Post
func (c *cURL) SendPOST(url string, params map[string]string) (map[string]interface{}, error) {
    res, err := c.client.Post(url, params)
    if err != nil {
        return make(map[string]interface{}), err
    }

    c.response = res.Body
    c.code = res.StatusCode
    result, err := c.Parser()
    if err != nil {
        return make(map[string]interface{}), err
    }
    return result, nil
}

// cURL.Post by json
func (c *cURL) SendPostByJson(url string, params interface{}) (map[string]interface{}, error) {
    res, err := c.client.PostJson(url, params)
    if err != nil {
        return make(map[string]interface{}), err
    }

    c.response = res.Body
    c.code = res.StatusCode
    result, err := c.Parser()
    if err != nil {
        return make(map[string]interface{}), err
    }
    return result, nil
}

// cURL.解析结果
func (c *cURL) Parser() (map[string]interface{}, error) {
    jsonMap := make(map[string]interface{})
    content, _ := ioutil.ReadAll(c.response)
    err := json.NewDecoder(bytes.NewReader(content)).Decode(&jsonMap)
    if err != nil {
        logger.Logger.Error(fmt.Sprintf("解析结果为：%s", string(content)))
        return make(map[string]interface{}), errors.Wrap(err, "解析返回结果失败")
    }
    return jsonMap, nil
}

// cURL.请求状态码
func (c *cURL) StatusCode() int {
    return c.code
}
