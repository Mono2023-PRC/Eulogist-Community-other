package RaknetConnection

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"

	_ "embed"
)

//go:embed skin_resource_patch.json
var skinResourcePatch []byte

//go:embed skin_geometry.json
var skinGeometry []byte

// 从 url 指定的网址下载文件，
// 并返回该文件的二进制形式
func DownloadFile(url string) (result []byte, err error) {
	// get http response
	httpResponse, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile: %v", err)
	}
	defer httpResponse.Body.Close()
	// read image data
	result, err = io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile: %v", err)
	}
	// return
	return
}

/*
从 url 指定的网址下载文件，
并处理为有效的皮肤数据。

skinImageData 指代皮肤的 PNG 二进制形式，
skinData 指代皮肤的一维的密集像素矩阵，
skinGeometryData 指代皮肤的骨架信息，
skinWidth 和 skinHight 则分别指代皮肤的
宽度和高度。

TODO: 支持 4D 皮肤
*/
func ProcessURLToSkin(url string) (
	skinImageData []byte,
	skinData []byte, skinGeometryData []byte,
	skinWidth int, skinHight int,
	err error,
) {
	// download skin file from remote server
	res, err := DownloadFile(url)
	if err != nil {
		return nil, nil, nil, 0, 0, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// get skin data
	if len(res) >= 4 && bytes.Equal(res[0:4], []byte("PK\x03\x04")) {
		// TODO: 支持 4D 皮肤
		{
			// skinImageData, skinGeometryData, err = ConvertZIPToSkin(res)
			skinImageData, _, err = ConvertZIPToSkin(res)
			if err != nil {
				return nil, nil, nil, 0, 0, fmt.Errorf("ProcessURLToSkin: %v", err)
			}
			skinGeometryData = skinGeometry
		}
	} else {
		skinImageData, skinGeometryData = res, skinGeometry
	}
	// decode to image
	img, err := ConvertToPNG(skinImageData)
	if err != nil {
		return nil, nil, nil, 0, 0, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// encode to pixels and return
	return skinImageData, img.(*image.NRGBA).Pix, skinGeometryData, img.Bounds().Dx(), img.Bounds().Dy(), nil
}

// 从 zipData 指代的 ZIP 二进制数据负载提取皮肤数据。
// skinImageData 代表皮肤的 PNG 二进制形式，
// skinGeometry 代表皮肤的骨架信息。
//
// TODO: 支持 4D 皮肤
func ConvertZIPToSkin(zipData []byte) (skinImageData []byte, skinGeometryData []byte, err error) {
	// prepare
	skinImageBuffer := bytes.NewBuffer([]byte{})
	skinGeometryBuffer := bytes.NewBuffer([]byte{})
	// create reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
	}
	// find skin contents
	for _, file := range reader.File {
		// skin data
		if strings.Contains(file.Name, ".png") {
			r, err := file.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer func() {
				err = r.Close()
				if err != nil {
					skinImageData, skinGeometryData, err = nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
				}
			}()
			_, err = io.Copy(skinImageBuffer, r)
			if err != nil {
				return nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
		}
		// skin geometry
		if strings.Contains(file.Name, "geometry.json") {
			r, err := file.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer func() {
				err = r.Close()
				if err != nil {
					skinImageData, skinGeometryData, err = nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
				}
			}()
			_, err = io.Copy(skinGeometryBuffer, r)
			if err != nil {
				return nil, nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
		}
	}
	// return
	return skinImageBuffer.Bytes(), skinGeometryBuffer.Bytes(), nil
}

// 将 imageData 解析为 PNG 图片
func ConvertToPNG(imageData []byte) (image.Image, error) {
	buffer := bytes.NewBuffer(imageData)
	img, err := png.Decode(buffer)
	if err != nil {
		return nil, fmt.Errorf("ConvertToPNG: %v", err)
	}
	return img, nil
}
