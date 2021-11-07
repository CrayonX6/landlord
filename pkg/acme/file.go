package acme

// export CGO_CFLAGS_ALLOW="-Xpreprocessor"

import (
    "bufio"
    "bytes"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/fogleman/gg"
    "golang.org/x/image/bmp"
    "image"
    "image/color"
    "image/draw"
    "image/gif"
    "image/jpeg"
    "image/png"
    "io/ioutil"
    "os"
    "path/filepath"
)

type Circle struct {
    P image.Point
    R int
}

var GIF = []byte("GIF")
var BMP = []byte("BM")
var JPG = []byte{0xff, 0xd8, 0xff}
var PNG = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

const (
    GifType = "image/gif"
    BmpType = "image/x-ms-bmp"
    JpgType = "image/jpeg"
    PngType = "image/png"
)

func (c *Circle) ColorModel() color.Model {
    return color.AlphaModel
}

func (c *Circle) Bounds() image.Rectangle {
    return image.Rect(c.P.X-c.R, c.P.Y-c.R, c.P.X+c.R, c.P.Y+c.R)
}

func (c *Circle) At(x, y int) color.Color {
    xx, yy, rr := float64(x-c.P.X)+0.5, float64(y-c.P.Y)+0.5, float64(c.R)
    if xx*xx+yy*yy < rr*rr {
        return color.Alpha{A: 255}
    }
    return color.Alpha{}
}

// 将图片变成圆形
func CircleImage(imageFile string, saveFile string, r int) error {
    imgFile, err := os.Open(imageFile)
    if err != nil {
        return err
    }
    defer func() {
        _ = imgFile.Close()
    }()

    img, err := png.Decode(imgFile)
    if err != nil {
        return err
    }

    rgba := image.NewRGBA(image.Rect(0, 0, r*2, r*2))
    draw.DrawMask(rgba, rgba.Bounds(), img, image.ZP, &Circle{P: image.Point{X: r, Y: r}, R: r}, image.ZP, draw.Over)
    err = RGBAToImage(saveFile, rgba)
    if err != nil {
        return err
    }
    return nil
}

// 图片转码
func DecodeImage(file string) (image.Image, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer func() {
        _ = f.Close()
    }()
    imageType, err := ImageType(file)
    if err != nil {
        return nil, err
    }

    switch imageType {
    case PngType:
        return png.Decode(f)
    case GifType:
        return gif.Decode(f)
    case BmpType:
        return bmp.Decode(f)
    case JpgType:
        return jpeg.Decode(f)
    default:
        return jpeg.Decode(f)
    }
}

// 获取图片类型
func ImageType(file string) (string, error) {
    f, err := ioutil.ReadFile(file)
    if err != nil {
        return "", err
    }
    var imageType string
    if bytes.Equal(PNG, f[0:8]) {
        imageType = PngType
    }
    if bytes.Equal(GIF, f[0:3]) {
        imageType = GifType
    }
    if bytes.Equal(BMP, f[0:2]) {
        imageType = BmpType
    }
    if bytes.Equal(JPG, f[0:3]) {
        imageType = JpgType
    }

    if imageType == "" {
        return imageType, errors.New("unknown image type")
    }
    return imageType, nil
}

// 将图片转换成 PNG
func ImageToPng(file string, saveFile string) error {
    save, err := os.Create(saveFile)
    if err != nil {
        return err
    }

    img, err := DecodeImage(file)
    if err != nil {
        return err
    }

    if err := png.Encode(save, img); err != nil {
        return err
    }
    return nil
}

// 保存 RGBA
func RGBAToImage(saveFile string, rgba image.Image) error {
    f, err := os.Create(saveFile)
    if err != nil {
        return err
    }
    defer func() {
        _ = f.Close()
    }()

    b := bufio.NewWriter(f)
    err = png.Encode(b, rgba)
    if err != nil {
        return err
    }

    err = b.Flush()
    if err != nil {
        return err
    }
    return nil
}

// 图片转 RGBA
func ImageToRGBA(src image.Image) *image.RGBA {
    if dst, ok := src.(*image.RGBA); ok {
        return dst
    }

    b := src.Bounds()
    dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
    draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
    return dst
}

// 创建矩形底图
func CreateRectBackground(saveFile string, width, height int, src image.Image) error {
    rgba := image.NewRGBA(image.Rect(0, 0, width, height))
    draw.Draw(rgba, rgba.Bounds(), src, image.Point{}, draw.Src)
    err := RGBAToImage(saveFile, rgba)
    if err != nil {
        return err
    }
    return nil
}

type TextWatermark struct {
    Text     string  `json:"text"`
    Font     string  `json:"font"`
    Size     float64 `json:"size"`
    Dissolve int     `json:"dissolve"`
    HexColor string  `json:"hex_color"`
    Left     float64 `json:"left"`
    Top      float64 `json:"top"`
}

// 图中添加水印文字
func CreateTextWatermark(bgFilename string, width, height int, saveFile string, quality int, options []TextWatermark) error {
    img, err := gg.LoadImage(bgFilename)
    if err != nil {
        return err
    }

    dc := gg.NewContext(width, height)
    dc.DrawImage(img, 0, 0)

    for _, opt := range options {
        if len(opt.Text) == 0 {
            continue
        }
        dc.SetHexColor(opt.HexColor)
        if err := dc.LoadFontFace(opt.Font, opt.Size); err != nil {
            return err
        }
        dc.DrawStringAnchored(opt.Text, opt.Left, opt.Top, 0, 0)
    }
    dc.Clip()
    _ = gg.SaveJPG(saveFile, dc.Image(), quality)
    return nil
}

// base64保存为图片
func Base64ToFile(saveFile, base64string string) (err error) {
    d, err := base64.StdEncoding.DecodeString(base64string)
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(saveFile, d, 0666)
    if err != nil {
        return err
    }
    return nil
}

// 图片转base64
func ImagesToBase64(imgFile string) []byte {
    ff, _ := os.Open(imgFile)
    fileInfo, _ := ff.Stat()
    defer func() {
        _ = ff.Close()
    }()
    buffer := make([]byte, fileInfo.Size())
    n, _ := ff.Read(buffer)
    base64string := base64.StdEncoding.EncodeToString(buffer[:n])
    return []byte(base64string)
}

// 创建目录
func MkDir(path string, isFile bool) error {
    if isFile {
        path, _ = filepath.Split(path)
    }
    err := os.MkdirAll(path, os.ModePerm)
    return err
}

// 递归目录
func DirectoryRecursion(
    path string,
    fileFn func(file string, info os.FileInfo) error,
    dirFn func(dir string, info os.FileInfo) error,
) error {
    return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if info == nil {
            // when filename is illegal then info == nil, if return err while be stop in current
            fmt.Printf("Can't find path %s", path)
            return nil
        }
        if err != nil {
            return err
        }
        if info.IsDir() {
            if dirFn != nil {
                return dirFn(path, info)
            }
        } else {
            if fileFn != nil {
                return fileFn(path, info)
            }
        }
        return nil
    })
}
