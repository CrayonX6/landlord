package acme

import (
    "bytes"
    "beginning/pkg/file"
    "context"
    "fmt"
    "github.com/spf13/viper"
    "io/ioutil"
)

type ThirdCOS struct {
}

// 获取 COS 域名
func (cc *ThirdCOS) GetCosHost(scheme interface{}, bucket interface{}, region interface{}) string {
    if scheme == nil {
        scheme = "https"
    }
    if bucket == nil {
        bucket = viper.GetString("fs.bucket")
    }
    if region == nil {
        region = viper.GetString("fs.region")
    }
    return fmt.Sprintf("%s://%s.cos.%s.myqcloud.com", scheme, bucket.(string), region.(string))
}

// 上传文件到 COS
func (cc *ThirdCOS) UploadToCOS(filename string, name string) (string, error) {
    buffer, err := ioutil.ReadFile(filename)
    if err != nil {
        return "", err
    }

    _, err = file.Fs.Object.Put(context.Background(), name, bytes.NewReader(buffer), nil)
    if err != nil {
        return "", err
    }

    // 生成地址
    url := fmt.Sprintf("%s/%s", cc.GetCosHost("https", nil, nil), name)
    return url, nil
}
