package acme

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/golang-module/carbon"
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
    "github.com/shopspring/decimal"
    "io"
    "io/ioutil"
    "math"
    "math/big"
    "math/rand"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "path/filepath"
    "reflect"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

// 打印结构体
func PrintVar(items ...interface{}) {
    if len(items) > 10 {
        fmt.Println(" -- 10 variables most at a moment -- ")
        return
    }
    split := "-=+*#@%!&~"
    for k, v := range items {
        if v == nil {
            fmt.Println()
            continue
        }
        variable := fmt.Sprintf(">> %#v", v)
        fmt.Println(" (" + strconv.Itoa(k+1) + ") " + strings.Repeat(split[k:k+1], 5) + variable)
    }
    fmt.Println()
}

// Md5 哈希值
func Md5(target string) string {
    hash := md5.New()
    _, _ = io.WriteString(hash, target)

    return fmt.Sprintf("%x", hash.Sum(nil))
}

// 获取变量的数据类型
func Typeof(v interface{}) string {
    return reflect.TypeOf(v).String()
}

// AES Padding
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
    padding := blockSize - len(cipherText)%blockSize
    padText := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(cipherText, padText...)
}

// AES UnPadding
func PKCS7UnPadding(origData []byte) []byte {
    length := len(origData)
    unPadding := int(origData[length-1])
    return origData[:(length - unPadding)]
}

// 微信解密 AES 密文
func WxAESDecrypt(encrypted string, key string, iv string) string {
    src, _ := base64.StdEncoding.DecodeString(encrypted)
    _key, _ := base64.StdEncoding.DecodeString(key)
    _iv, _ := base64.StdEncoding.DecodeString(iv)

    block, _ := aes.NewCipher(_key)
    mode := cipher.NewCBCDecrypter(block, _iv)
    dst := make([]byte, len(src))
    mode.CryptBlocks(dst, src)
    dst = PKCS7UnPadding(dst)

    return string(dst)
}

// 判断一个值是否在切片中
func InArray(need interface{}, haystack interface{}) (exists bool, index int) {
    exists = false
    index = -1
    switch reflect.TypeOf(haystack).Kind() {
    case reflect.Slice:
        s := reflect.ValueOf(haystack)
        for i := 0; i < s.Len(); i++ {
            if reflect.DeepEqual(need, s.Index(i).Interface()) == true {
                index = i
                exists = true
                return
            }
        }
    }
    return
}

// 获取今天是周几
func TodayIsWeekTh(now carbon.Carbon) string {
    weekMap := map[int]string{
        0: "周日",
        1: "周一",
        2: "周二",
        3: "周三",
        4: "周四",
        5: "周五",
        6: "周六",
        7: "周日",
    }
    return weekMap[now.DayOfWeek()]
}

// 生成随机数
func RandInt(min, max int64) int64 {
    if min > max {
        min, max = max, min
    }
    return rand.Int63n(max-min+1) + min
}

// 生成随机字符串
func RandStr(digit int) (container string) {
    dict := Dict4Decimal4("")
    dictLen := int64(len([]rune(dict)))
    for i := 0; i < digit; i++ {
        container += string(dict[RandInt(0, dictLen-1)])
    }
    return container
}

// 截取字符串
//
// start
//   正数    在字符串的指定位置开始,超出字符串长度强制把start变为字符串长度
//   负数    在从字符串结尾的指定位置开始
//   0        在字符串中的第一个字符处开始
// length
//   正数    从 start 参数所在的位置返回
//   负数     从字符串末端返回
func Substr(str string, start, length int) string {
    if length == 0 {
        return ""
    }
    runeStr := []rune(str)
    lenStr := len(runeStr)

    if start < 0 {
        start = lenStr + start
    }
    if start > lenStr {
        start = lenStr
    }
    end := start + length
    if end > lenStr {
        end = lenStr
    }
    if length < 0 {
        end = lenStr + length
    }
    if start > end {
        start, end = end, start
    }
    return string(runeStr[start:end])
}

// 今日剩余秒数
func TodaySurplusSecond() int64 {
    now := carbon.Now()
    surplus := now.EndOfDay().ToTimestamp() - now.ToTimestamp()
    return surplus
}

// 下载远程文件到服务器
func DownloadRemoteFile(fileUrl, filePath string, retryTimes int) error {
    if retryTimes <= 0 {
        return errors.New(fmt.Sprintf("多次重试下载失败, 文件：%s -> %s", fileUrl, filePath))
    }
    response, err := http.Get(fileUrl)
    if err != nil {
        return err
    }
    defer func() {
        _ = response.Body.Close()
    }()

    if response.StatusCode == 404 {
        return errors.New(fmt.Sprintf("下载的文件不存在, 文件：%s", fileUrl))
    }

    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer func() {
        _ = file.Close()
    }()

    source, _ := ioutil.ReadAll(response.Body)
    _, err = file.Write(source)
    if err != nil {
        return err
    }

    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return err
    }
    if len(content) == 0 {
        PrintVar(fmt.Sprintf("文件下载后大小为：%d，进行倒数第%d次重试 [%s -> %s]", len(content), retryTimes, fileUrl, filePath))
        time.Sleep(time.Millisecond * 100)
        return DownloadRemoteFile(fileUrl, filePath, retryTimes-1)
    }

    return nil
}

// 监测文件是否下载完成
func IsDownloadDown(file string, doneSize int, retryTimes int) error {
    if retryTimes <= 0 {
        return errors.New(fmt.Sprintf("监测超过最大次数后仍未下载完, 文件：%s", file))
    }
    content, err := ioutil.ReadFile(file)
    if err != nil {
        return err
    }
    if len(content) < doneSize-50 {
        time.Sleep(time.Millisecond * 100)
        return IsDownloadDown(file, doneSize, retryTimes-1)
    }
    return nil
}

// 复制文件
func CopyFile(src, des string) (written int64, err error) {
    if src == des {
        return
    }
    srcFile, err := os.Open(src)
    if err != nil {
        return 0, err
    }
    defer func() {
        _ = srcFile.Close()
    }()

    // 获取源文件的权限
    fi, _ := srcFile.Stat()
    perm := fi.Mode()

    desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) // 复制源文件的所有权限
    if err != nil {
        return 0, err
    }
    defer func() {
        _ = desFile.Close()
    }()

    return io.Copy(desFile, srcFile)
}

// 最小数
func Min(member ...int) int {
    var min int
    for i, m := range member {
        if i == 0 {
            min = m
            continue
        }
        if m < min {
            min = m
        }
    }
    return min
}

// 最大数
func Max(member ...int) int {
    var max int
    for i, m := range member {
        if i == 0 {
            max = m
            continue
        }
        if m > max {
            max = m
        }
    }
    return max
}

// 耗时统计
func Cost(ck string, label string, prevPointNano int64) int64 {
    if len(ck) == 0 {
        return 0
    }
    cost, now := GetCost(prevPointNano)
    if len(label) > 0 {
        fmt.Printf(" [%s] >>> 耗时：%d \t%s\n", ck, cost, label)
    }
    return now
}

// 获取耗时统计
func GetCost(prevPointNano int64) (int64, int64) {
    now := time.Now().UnixNano() / 1e6
    cost := now - prevPointNano
    if prevPointNano == 0 {
        cost = 0
    }
    return cost, now
}

// 安全版base64编码
func SafeBase64Encode(src string) (dest string, err error) {
    dest = base64.StdEncoding.EncodeToString([]byte(src))
    dest = strings.ReplaceAll(dest, "+", "-")
    dest = strings.ReplaceAll(dest, "/", "_")
    return
}

// 安全版base64解码
func SafeBase64Decode(src string) (dest string, err error) {
    src = strings.ReplaceAll(src, "-", "+")
    src = strings.ReplaceAll(src, "_", "/")
    d, err := base64.StdEncoding.DecodeString(src)
    if err != nil {
        return
    }
    dest = string(d)
    return
}

// 正则替换
func RegReplace(text string, regStr string, handler func(args ...string) string) string {
    reg := regexp.MustCompile(regStr)
    match := reg.FindAllStringSubmatch(text, -1)
    for _, chunkArr := range match {
        chunk := chunkArr[0]
        chunkNew := handler(chunkArr[1:]...)
        text = strings.Replace(text, chunk, chunkNew, 1)
    }
    return text
}

// URL追加查询参数
func UrlAppendQuery(url string, field string, value interface{}) string {
    if strings.Contains(url, "?") {
        url = fmt.Sprintf("%s&%s=%s", url, field, ToStr(value))
    } else {
        url = fmt.Sprintf("%s?%s=%s", url, field, ToStr(value))
    }
    return url
}

// 解析 url 参数到 domain
func ParseUrlDomain(target string) string {
    u, _ := url.Parse(target)
    var port string
    if exists, _ := InArray(u.Port(), []string{"80", "443"}); !exists {
        port = fmt.Sprintf(":%s", u.Port())
    }
    return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, port)
}

// 解析 url 参数到 map
func ParseUrlQuery(target string) map[string]interface{} {
    var args = make(map[string]interface{})
    u, _ := url.Parse(target)
    for k, v := range u.Query() {
        if len(v) == 0 {
            args[k] = nil
        } else if len(v) == 1 {
            args[k] = v[0]
        } else {
            args[k] = v
        }
    }
    return args
}

// 将 map 转查询字符串
func MapToQueryString(target map[string]interface{}, queryField string, kvSplit string, pairSplit string) string {
    var keys []string
    for k := range target {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    var query []string
    for _, k := range keys {
        var chunk string
        var v = target[k]
        var field = k
        if len(queryField) > 0 {
            field = fmt.Sprintf("%s[%s]", queryField, k)
        }
        if IsMap(v) {
            chunk = MapToQueryString(v.(map[string]interface{}), field, kvSplit, pairSplit)
        } else if IsArray(v) || IsSlice(v) {
            chunk = ArrayToQueryString(v.([]interface{}), field, kvSplit, pairSplit)
        } else {
            chunk = fmt.Sprintf("%s%s%s", field, kvSplit, ToStr(v))
        }
        query = append(query, chunk)
    }
    return strings.Join(query, pairSplit)
}

// 将 array 转查询字符串
func ArrayToQueryString(target []interface{}, queryField string, kvSplit string, pairSplit string) string {
    var query []string
    for k, v := range target {
        var chunk string
        var field = fmt.Sprintf("%s[%d]", queryField, k)
        if IsMap(v) {
            chunk = MapToQueryString(v.(map[string]interface{}), field, kvSplit, pairSplit)
        } else if IsArray(v) || IsSlice(v) {
            chunk = ArrayToQueryString(v.([]interface{}), field, kvSplit, pairSplit)
        } else {
            chunk = fmt.Sprintf("%s%s%s", field, kvSplit, ToStr(v))
        }
        query = append(query, chunk)
    }
    return strings.Join(query, pairSplit)
}

// 将 interface 转查询字符串
func InterfaceToQueryString(target interface{}, kvSplit string, pairSplit string) string {
    var targetMap map[string]interface{}
    AlignStructAndMap(target, &targetMap)
    return MapToQueryString(targetMap, "", kvSplit, pairSplit)
}

// 对 map 签名
func MapToSignature(target map[string]interface{}, salt string, timeKey string, kvSplit string, pairSplit string, saltSplit string) (signStr string, signMd5 string) {
    if _, exists := target[timeKey]; !exists {
        target[timeKey] = carbon.Now().ToTimestampWithMicrosecond()
    }
    signature := MapToQueryString(target, "", kvSplit, pairSplit)
    signature = fmt.Sprintf("%s%s%s", signature, saltSplit, salt)
    return signature, Md5(signature)
}

// 对 interface 签名
func InterfaceToSignature(target interface{}, salt string, timeKey string, kvSplit string, pairSplit string, saltSplit string) (signStr string, signMd5 string, targetMap map[string]interface{}) {
    AlignStructAndMap(target, &targetMap)
    signStr, signMd5 = MapToSignature(targetMap, salt, timeKey, kvSplit, pairSplit, saltSplit)
    return
}

// 删除 map 指定下标的成员
func DigMapField(target map[string]interface{}, field string) (map[string]interface{}, interface{}) {
    var val interface{}
    if _, exists := target[field]; exists {
        val = target[field]
        delete(target, field)
    }
    return target, val
}

// 删除数组指定下标的成员
func DigIntArrayIndex(target []int64, index int64) ([]int64, int64) {
    var val int64
    if int64(len(target)) > index {
        val = target[index]
        target = append(target[:index], target[index+1:]...)
    }
    return target, val
}

// 对二维 map 数组排序
func SortMapArray(target []map[string]interface{}, field string, mode string) []map[string]interface{} {
    var newTarget = make(map[string]map[string]interface{})
    var keys []string
    for k, v := range target {
        val := ToStr(v[field])
        val = fmt.Sprintf("%s_%d", val, k)
        newTarget[val] = v
        keys = append(keys, val)
    }
    mode = strings.ToUpper(mode)
    if mode == "ASC" {
        sort.Strings(keys)
    } else if mode == "DESC" {
        sort.Sort(sort.Reverse(sort.StringSlice(keys)))
    }
    var result []map[string]interface{}
    for _, val := range keys {
        result = append(result, newTarget[val])
    }
    return result
}

// 对数组排序
func SortArrayByString(target []interface{}, mode string, handler func(item interface{}) string) []interface{} {
    var newTarget = make(map[string]interface{})
    var keys []string
    for k, v := range target {
        val := handler(v)
        val = fmt.Sprintf("%s_%d", val, k)
        newTarget[val] = v
        keys = append(keys, val)
    }
    mode = strings.ToUpper(mode)
    if mode == "ASC" {
        sort.Strings(keys)
    } else if mode == "DESC" {
        sort.Sort(sort.Reverse(sort.StringSlice(keys)))
    }
    var result []interface{}
    for _, val := range keys {
        result = append(result, newTarget[val])
    }
    return result
}

// 对数组排序
func SortArrayByFloat(target []interface{}, mode string, handler func(item interface{}) float64) []interface{} {
    var newTarget = make(map[float64]interface{})
    var keys []float64
    for k, v := range target {
        val := handler(v)
        val, _ = decimal.NewFromFloat(val).Add(decimal.NewFromFloat(float64(k) / 1000000)).Float64()
        newTarget[val] = v
        keys = append(keys, val)
    }
    mode = strings.ToUpper(mode)
    if mode == "ASC" {
        sort.Float64s(keys)
    } else if mode == "DESC" {
        sort.Sort(sort.Reverse(sort.Float64Slice(keys)))
    }
    var result []interface{}
    for _, val := range keys {
        result = append(result, newTarget[val])
    }
    return result
}



// 补齐日期为索引的数组
func PerfectDateArray(beginDay carbon.Carbon, endDay carbon.Carbon, dateField string, sortMode string, callback func(date carbon.Carbon) map[string]interface{}) (list []map[string]interface{}) {
    if beginDay.DiffInSeconds(endDay) < 86400 {
        return
    }
    for {
        item := callback(beginDay)
        list = append(list, item)
        beginDay = beginDay.AddDay()
        if beginDay.Gte(endDay) {
            break
        }
    }
    return
}

// 补齐日期为索引的 Map
func PerfectDateMap(beginDay carbon.Carbon, endDay carbon.Carbon, dateField string, callback func(date carbon.Carbon) map[string]interface{}) map[string]map[string]interface{} {
    list := make(map[string]map[string]interface{})
    if beginDay.DiffInSeconds(endDay) < 1 {
        return list
    }
    for {
        item := callback(beginDay)
        list[beginDay.ToDateString()] = item
        beginDay = beginDay.AddDay()
        if beginDay.Gt(endDay) {
            break
        }
    }
    return list
}

// Map 转 Array
func MapToArray(source map[string]map[string]interface{}, field string) (target []map[string]interface{}) {
    for key, item := range source {
        if len(field) > 0 {
            item[field] = key
        }
        target = append(target, item)
    }
    return
}

// Array 转 Map
func ArrayToMap(source []map[string]interface{}, field string) map[string]map[string]interface{} {
    target := make(map[string]map[string]interface{})
    for _, item := range source {
        target[item[field].(string)] = item
    }
    return target
}

// 字段升序排列SQL（为空或0则排在最后）
func SortByFieldAsc(field string, asSortField string) string {
    return fmt.Sprintf("CASE WHEN %s = 0 OR %s IS NULL THEN %d ELSE %s END AS %s", field, field, math.MaxInt32, field, asSortField)
}

// 字段降序排列SQL（为空或0则排在最后）
func SortByFieldDesc(field string, asSortField string) string {
    return fmt.Sprintf("CASE WHEN %s = 0 OR %s IS NULL THEN %d ELSE %s END AS %s", field, field, 0, field, asSortField)
}

// 特殊字符串
func Special4char(extra string) string {
    return fmt.Sprintf("`-=[];'\\,.//~!@#$%^&*()_+{}:\"|<>?·【】；’、，。、！￥…（）—：“《》？%s", extra)
}

// 进制转换字典(数字、小写字母、大写字母)
func Dict4Decimal1(extra string) string {
    return fmt.Sprintf("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ%s", extra)
}

// 进制转换字典(小写字母、大写字母)
func Dict4Decimal2(extra string) string {
    return fmt.Sprintf("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ%s", extra)
}

// 进制转换字典(数字、小写字母)
func Dict4Decimal3(extra string) string {
    return fmt.Sprintf("0123456789abcdefghijklmnopqrstuvwxyz%s", extra)
}

// 进制转换字典(数字、大写字母)
func Dict4Decimal4(extra string) string {
    return fmt.Sprintf("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ%s", extra)
}

// 进制转换字典(数字、小写字母、大写字母)乱序
func Dict4Decimal5(extra string) string {
    return fmt.Sprintf("AWi2QFN3VqUC4xPDazgXEOut1feMLdTbHK9sZrRJv5j7pcy8SkmYl60oBwIGnh%s", extra)
}

// 分割字符串
func SplitString(source string) (t1 []string, t2 map[string]int64) {
    t1 = strings.Split(source, "")
    t2 = make(map[string]int64)
    for index, char := range t1 {
        t2[char] = int64(index)
    }
    return
}

// 十进制转N进制
func Decimal2N(sourceNumber int64, N int, extra string) string {
    newNumStr := ""
    var remainderString string
    t1, _ := SplitString(Dict4Decimal1(extra))
    if N > len(t1) {
        panic(fmt.Sprintf("Argument N must lte the length (%d) of dict", len(t1)))
    }
    for sourceNumber != 0 {
        remainderString = t1[sourceNumber%int64(N)]
        newNumStr = remainderString + newNumStr
        sourceNumber = sourceNumber / int64(N)
    }
    return newNumStr
}

// N进制转十进制
func N2Decimal(sourceString string, N int, extra string) int {
    var newNum = 0.0
    nNum := len(strings.Split(sourceString, "")) - 1
    t1, t2 := SplitString(Dict4Decimal1(extra))
    if N > len(t1) {
        panic(fmt.Sprintf("Argument N must lte the length (%d) of dict", len(t1)))
    }
    for _, value := range strings.Split(sourceString, "") {
        tmp := float64(t2[value])
        if tmp == -1 {
            break
        }
        newNum = newNum + tmp*math.Pow(float64(N), float64(nNum))
        nNum = nNum - 1
    }
    return int(newNum)
}

// 拷贝文件夹
func CopyDir(srcPath string, destPath string) error {
    if srcInfo, err := os.Stat(srcPath); err != nil {
        return err
    } else {
        if !srcInfo.IsDir() {
            err := errors.New("source path is not directory")
            return err
        }
    }
    if destInfo, err := os.Stat(destPath); err != nil {
        return err
    } else {
        if !destInfo.IsDir() {
            err := errors.New("target path is not directory")
            return err
        }
    }
    err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
        if f == nil {
            return err
        }
        if !f.IsDir() {
            path := strings.Replace(path, "\\", "/", -1)
            destNewPath := strings.Replace(path, srcPath, destPath, -1)
            _, _ = CopyFile(path, destNewPath)
        }
        return nil
    })
    return err
}

// 系统命令拷贝方式
func CopyFileByCmd(src, des string) ([]byte, error) {
    if src == des {
        return []byte{}, nil
    }
    cmd := exec.Command("cp", src, des)
    return cmd.Output()
}

// 检测目录是否存在
func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

// 文件后缀
func FileSuffix(file string) string {
    if !strings.Contains(file, ".") {
        return ""
    }
    items := strings.Split(file, ".")
    return items[len(items)-1]
}

// 文件名
func FileName(file string) string {
    if !strings.Contains(file, "/") {
        return file
    }
    items := strings.Split(file, "/")
    return items[len(items)-1]
}

// 粗略计算文本长度
func TextWidth(text string, len4byte1 float64, len4byte3 float64) float64 {
    var length float64
    for _, v := range []rune(text) {
        if len([]byte(string(v))) == 3 {
            length += len4byte3
        } else {
            length += len4byte1
        }
    }
    return length
}

// 每N个字符分割字符串
func SplitPerChar(target string, per int) []string {
    var result []string
    var page int
    for {
        part := Substr(target, page*per, per)
        if len(part) == 0 {
            break
        }
        result = append(result, part)
        page += 1
    }
    return result
}

// 每N个字符插入指定字符（串）
func InsertXPerChar(target string, per int, x string) string {
    items := SplitPerChar(target, per)
    result := strings.Join(items, x)
    return result
}

// 大整数加工
func BigIntHandler(binIntNumber string) (number *big.Int) {
    if len(binIntNumber) == 0 {
        binIntNumber = "0"
    }
    number, _ = new(big.Int).SetString(binIntNumber, 10)
    return
}

// 大浮点数加工
func BigFloatHandler(binFloatNumber string) (number *big.Float) {
    if len(binFloatNumber) == 0 {
        binFloatNumber = "0"
    }
    number, _ = new(big.Float).SetString(binFloatNumber)
    return
}

// 大整数加法
func BigIntAdd(leftInt string, rightInt string) (intResult string) {
    l := BigIntHandler(leftInt)
    r := BigIntHandler(rightInt)
    intResult = l.Add(l, r).String()
    return
}

// 大整数减法
func BigIntSub(leftInt string, rightInt string) (intResult string, gteZero bool) {
    l := BigIntHandler(leftInt)
    r := BigIntHandler(rightInt)
    intResult = l.Sub(l, r).String()
    gteZero = strings.Index(intResult, "-") != 0
    return
}

// 大整数比较 <=>
// 小于、等于、大于 得到的结果一次为 -1、0、1
func BigIntCompare(leftInt string, rightInt string) int {
    l := BigIntHandler(leftInt)
    r := BigIntHandler(rightInt)
    return l.Cmp(r)
}

// 大整数乘法
func BigIntMul(leftInt string, rightInt string) (intResult string) {
    l := BigIntHandler(leftInt)
    r := BigIntHandler(rightInt)
    intResult = l.Mul(l, r).String()
    return
}

// 大整数除法
func BigIntDiv(leftInt string, rightInt string) (intResult string) {
    l := BigIntHandler(leftInt)
    r := BigIntHandler(rightInt)
    intResult = l.Div(l, r).String()
    return
}

// 大浮点数乘法 (向上取整)
func BigFloatMul(leftFloat string, rightFloat string) (intResult string) {
    l := BigFloatHandler(leftFloat)
    r := BigFloatHandler(rightFloat)
    floatValue := l.Mul(l, r)
    if floatValue.IsInt() { // 刚好为整数
        intValue, _ := floatValue.Int(nil)
        intResult = intValue.String()
    } else { // 向上取整
        o := BigIntHandler("1")
        i, _ := floatValue.Int(nil)
        intResult = i.Add(i, o).String()
    }
    return
}

// 大浮点数除法
func BigFloatDiv(leftFloat string, rightFloat string) (intResult string) {
    l := BigFloatHandler(leftFloat)
    r := BigFloatHandler(rightFloat)
    floatValue := l.Quo(l, r)
    if floatValue.IsInt() { // 刚好为整数
        intValue, _ := floatValue.Int(nil)
        intResult = intValue.String()
    } else { // 向上取整
        o := BigIntHandler("1")
        i, _ := floatValue.Int(nil)
        intResult = i.Add(i, o).String()
    }
    return
}

// 数组中的最大值和最小值
func BoundaryOfIntArray(target []int64, notZero bool) (minK int, minV int64, maxK int, maxV int64) {
    minK = 0
    minV = target[minK]
    maxK = 0
    maxV = target[maxK]
    for i := 1; i < len(target); i++ {
        if notZero && target[i] == 0 {
            continue
        }
        if target[i] > maxV {
            maxK = i
            maxV = target[i]
        }
        if target[i] < minV {
            minK = i
            minV = target[i]
        }
    }
    return
}

// 数值边界截断
func BoundaryTruncate(number int64, min int64, max int64) int64 {
    if number < min {
        return min
    }
    if number > max {
        return max
    }
    return number
}

// 整数数组转字符串
func JoinInt64Array(target []int64, sep string) (latest string) {
    for _, val := range target {
        valStr := ToStr(val)
        latest = fmt.Sprintf("%s%s%s", latest, sep, valStr)
    }
    latest = strings.Trim(latest, sep)
    return
}

// 计算两个时间段的交集秒数
func CalIntersectSeconds(leftBegin int64, leftEnd int64, rightBegin int64, rightEnd int64) (from int64, to int64, err error) {
    if leftBegin >= leftEnd || rightBegin >= rightEnd {
        err = errors.New("开始时间不能大于结束时间")
        return
    }
    if leftEnd <= rightBegin || leftBegin >= rightEnd {
        err = errors.New("两个时间段无交集")
        return
    }

    item := []int{int(leftBegin), int(leftEnd), int(rightBegin), int(rightEnd)}
    sort.Sort(sort.IntSlice(item))

    from = int64(item[1])
    to = int64(item[2])
    return
}

// 浮点数保留N位小数 (四舍五入)
func FloatToFixed(value float64, decimal int) float64 {
    format := fmt.Sprintf("%%.%df", decimal)
    value, _ = strconv.ParseFloat(fmt.Sprintf(format, value), 64)
    return value
}

// 谷歌验证码
func GoogleAuthenticatorCode(name string, secret string, issuer string) (code string, content string) {
    content = fmt.Sprintf("otpauth://totp/%s?secret=%s", name, secret)
    if len(issuer) > 0 {
        content = fmt.Sprintf("%s&issur=%s", content, issuer)
    }

    key, _ := otp.NewKeyFromURL(content)
    code, _ = totp.GenerateCode(key.Secret(), carbon.Now().Time)
    return
}
