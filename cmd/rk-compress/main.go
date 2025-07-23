package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/gzip"
)

// # 启用版本控制模式
// ./compress -versioning -path ./Build

// # 普通模式（不生成版本文件）
// ./compress -path ./Build

// ./compress.exe -br 9 -gz 9 -workers 8 -path "Build"

type CompressTask struct {
	filePath string
	fileInfo os.FileInfo
}

var (
	versionData = make(map[string]string)
	versionLock sync.Mutex
)

func main() {
	brQuality := flag.Int("br", 6, "Brotli压缩级别(0-11)")
	gzLevel := flag.Int("gz", 7, "Gzip压缩级别(1-9)")
	workers := flag.Int("workers", runtime.NumCPU(), "并行worker数量")
	buildPath := flag.String("path", "./Build", "构建目录路径")
	versioning := flag.Bool("versioning", true, "启用版本号(MD5哈希)")
	flag.Parse()

	tasks := make(chan CompressTask, 1000)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go compressWorker(tasks, &wg, *brQuality, *gzLevel, *versioning, *buildPath)
	}

	err := filepath.Walk(*buildPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isTargetFile(path) {
			tasks <- CompressTask{
				filePath: path,
				fileInfo: info,
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("遍历目录错误: %v\n", err)
		return
	}

	close(tasks)
	wg.Wait()

	if *versioning {
		if err := generateVersionFile(*buildPath); err != nil {
			fmt.Printf("生成版本文件失败: %v\n", err)
		}
	}

	fmt.Println("压缩完成!")
	fmt.Println("按回车键退出...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func generateVersionFile(buildPath string) error {
	versionPath := filepath.Join(buildPath, "version.json")
	file, err := os.Create(versionPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(versionData)
}

func isTargetFile(filePath string) bool {
	switch {
	case strings.HasSuffix(filePath, ".data"),
		strings.HasSuffix(filePath, ".wasm"),
		strings.HasSuffix(filePath, "framework.js"),
		strings.HasSuffix(filePath, "symbols.json"):
		return true
	default:
		return false
	}
}

func compressWorker(tasks chan CompressTask, wg *sync.WaitGroup, brQuality, gzLevel int, versioning bool, buildPath string) {
	defer wg.Done()

	for task := range tasks {
		input, err := os.ReadFile(task.filePath)
		if err != nil {
			fmt.Printf("读取文件失败 %s: %v\n", task.filePath, err)
			continue
		}

		// 生成原始压缩文件（始终保留）
		baseBrPath := task.filePath + ".br"
		baseGzPath := task.filePath + ".gz"
		if err := compressBrotli(input, baseBrPath, brQuality); err != nil {
			fmt.Printf("Brotli失败 %s: %v\n", task.filePath, err)
		} else {
			fmt.Printf("生成原始Brotli: %s\n", baseBrPath)
		}
		if err := compressGzip(input, baseGzPath, gzLevel); err != nil {
			fmt.Printf("Gzip失败 %s: %v\n", task.filePath, err)
		} else {
			fmt.Printf("生成原始Gzip: %s\n", baseGzPath)
		}

		// 版本化处理
		if versioning {
			dir := filepath.Dir(task.filePath)
			filename := filepath.Base(task.filePath)
			ext := filepath.Ext(filename)
			basename := strings.TrimSuffix(filename, ext)
			hash := fmt.Sprintf("%x", md5.Sum(input))

			// 清理旧版本文件
			cleanOldVersions(dir, basename, ext)

			// 生成版本化文件名
			versionedName := fmt.Sprintf("%s.%s%s", basename, hash, ext)
			versionedBrPath := filepath.Join(dir, versionedName+".br")
			versionedGzPath := filepath.Join(dir, versionedName+".gz")

			// 记录版本信息
			relPath, _ := filepath.Rel(buildPath, task.filePath)
			versionLock.Lock()
			versionData[relPath] = versionedName
			versionLock.Unlock()

			// 生成版本化压缩文件
			if err := compressBrotli(input, versionedBrPath, brQuality); err != nil {
				fmt.Printf("版本Brotli失败 %s: %v\n", task.filePath, err)
			} else {
				fmt.Printf("生成版本Brotli: %s\n", versionedBrPath)
			}
			if err := compressGzip(input, versionedGzPath, gzLevel); err != nil {
				fmt.Printf("版本Gzip失败 %s: %v\n", task.filePath, err)
			} else {
				fmt.Printf("生成版本Gzip: %s\n", versionedGzPath)
			}
		}
	}
}

func cleanOldVersions(dir, basename, ext string) {
	patterns := []string{
		filepath.Join(dir, basename+".*"+ext+".br"),
		filepath.Join(dir, basename+".*"+ext+".gz"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			if err := os.Remove(match); err == nil {
				fmt.Printf("清理旧版本: %s\n", match)
			}
		}
	}
}

func compressBrotli(input []byte, outputPath string, quality int) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := brotli.NewWriterLevel(outFile, quality)
	defer writer.Close()

	_, err = writer.Write(input)
	return err
}

func compressGzip(input []byte, outputPath string, level int) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer, err := gzip.NewWriterLevel(outFile, level)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(input)
	return err
}
