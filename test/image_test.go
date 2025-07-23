package test

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/nfnt/resize"
)

// http://localhost:8080/public/robot.html
// go test -v -run TestRobt
func TestImage(t *testing.T) {
	srcImagePath := "big_image.jpeg" // 大图文件路径
	// tileSize := uint(100)           // 每个块的像素大小
	thumbWidth := uint(200)  // 缩略图宽度
	thumbHeight := uint(200) // 缩略图高度

	// // 执行切割图片为小块
	err := SliceImage(srcImagePath, 256)
	if err != nil {
		fmt.Printf("切割图片为小块出现错误: %s\n", err)
		return
	}

	// 生成原始大图的缩略图
	err = CreateThumbnail(srcImagePath, thumbWidth, thumbHeight)
	if err != nil {
		fmt.Printf("生成缩略图出现错误: %s\n", err)
		return
	}
}

// SliceImage 切分图片为多个小块，不足的部分用黑色填充
func SliceImage(srcImagePath string, tileSize uint) error {
	file, err := os.Open(srcImagePath)
	if err != nil {
		return fmt.Errorf("打开源图像失败: %w", err)
	}
	defer file.Close()

	srcImage, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码源图像失败: %w", err)
	}

	srcBounds := srcImage.Bounds()
	tileBounds := image.Rect(0, 0, int(tileSize), int(tileSize))

	for x := srcBounds.Min.X; x < srcBounds.Max.X; x += int(tileSize) {
		for y := srcBounds.Min.Y; y < srcBounds.Max.Y; y += int(tileSize) {
			subRect := image.Rect(x, y, x+int(tileSize), y+int(tileSize)).Intersect(srcBounds)
			subImg := image.NewRGBA(tileBounds)
			draw.Draw(subImg, tileBounds, &image.Uniform{color.Black}, image.Point{}, draw.Src)
			draw.Draw(subImg, subRect.Sub(image.Point{X: x, Y: y}), srcImage, subRect.Min, draw.Src)

			tilePath := filepath.Join("tiles", fmt.Sprintf("tile_%d_%d.jpg", y/int(tileSize), x/int(tileSize)))
			tileFile, err := os.Create(tilePath)
			if err != nil {
				return fmt.Errorf("创建切块文件失败: %w", err)
			}

			err = jpeg.Encode(tileFile, subImg, nil)
			if err != nil {
				tileFile.Close()
				return fmt.Errorf("保存切块文件失败: %w", err)
			}
			tileFile.Close()
		}
	}

	return nil
}

// CreateThumbnail 生成大图的缩略图
// srcImagePath: 原始大图路径
// thumbWidth: 缩略图宽度
// thumbHeight: 缩略图高度
func CreateThumbnail(srcImagePath string, thumbWidth, thumbHeight uint) error {
	file, err := os.Open(srcImagePath)
	if err != nil {
		return fmt.Errorf("打开源图像失败: %w", err)
	}
	defer file.Close()

	srcImage, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码源图像失败: %w", err)
	}

	thumbImg := resize.Resize(thumbWidth, thumbHeight, srcImage, resize.Lanczos3)
	thumbPath := filepath.Join("thumbnails", "thumbnail.jpg")
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("创建缩略图文件失败: %w", err)
	}
	defer thumbFile.Close()

	err = jpeg.Encode(thumbFile, thumbImg, nil)
	if err != nil {
		return fmt.Errorf("保存缩略图失败: %w", err)
	}

	return nil
}
