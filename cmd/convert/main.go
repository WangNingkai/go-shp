// Package main provides a command-line tool for converting between Shapefile and GeoJSON formats.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangningkai/go-shp"
)

func main() {
	var (
		input     = flag.String("input", "", "输入文件路径 (.shp 或 .geojson)")
		output    = flag.String("output", "", "输出文件路径")
		batch     = flag.Bool("batch", false, "批量转换模式")
		inputDir  = flag.String("input-dir", "", "批量转换输入目录")
		outputDir = flag.String("output-dir", "", "批量转换输出目录")
		help      = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *batch {
		handleBatchConversion(*inputDir, *outputDir)
		return
	}

	if *input == "" {
		fmt.Println("错误：必须指定输入文件")
		printHelp()
		os.Exit(1)
	}

	handleSingleConversion(*input, *output)
}

func printHelp() {
	fmt.Println("GeoJSON 和 Shapefile 转换工具")
	fmt.Println()
	fmt.Println("用法：")
	fmt.Println("  单文件转换：")
	fmt.Println("    go run cmd/convert/main.go -input=input.shp -output=output.geojson")
	fmt.Println("    go run cmd/convert/main.go -input=input.geojson -output=output.shp")
	fmt.Println()
	fmt.Println("  批量转换：")
	fmt.Println("    go run cmd/convert/main.go -batch -input-dir=./shapefiles -output-dir=./geojson")
	fmt.Println()
	fmt.Println("参数：")
	fmt.Println("  -input string")
	fmt.Println("        输入文件路径 (.shp 或 .geojson)")
	fmt.Println("  -output string")
	fmt.Println("        输出文件路径（可选，自动推断）")
	fmt.Println("  -batch")
	fmt.Println("        批量转换模式")
	fmt.Println("  -input-dir string")
	fmt.Println("        批量转换输入目录")
	fmt.Println("  -output-dir string")
	fmt.Println("        批量转换输出目录")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
}

func handleSingleConversion(input, output string) {
	ext := strings.ToLower(filepath.Ext(input))

	if output == "" {
		// 自动推断输出文件名
		base := strings.TrimSuffix(input, filepath.Ext(input))
		switch ext {
		case ".shp":
			output = base + ".geojson"
		case ".geojson":
			output = base + ".shp"
		default:
			log.Fatalf("不支持的文件类型：%s", ext)
		}
	}

	fmt.Printf("转换 %s 到 %s...\n", input, output)

	var err error
	switch ext {
	case ".shp":
		err = shp.ConvertShapefileToGeoJSON(input, output)
	case ".geojson":
		err = shp.ConvertGeoJSONToShapefile(input, output)
	default:
		log.Fatalf("不支持的输入文件类型：%s", ext)
	}

	if err != nil {
		log.Fatalf("转换失败：%v", err)
	}

	fmt.Printf("转换成功：%s\n", output)
}

func handleBatchConversion(inputDir, outputDir string) {
	if inputDir == "" || outputDir == "" {
		fmt.Println("错误：批量转换需要指定输入和输出目录")
		printHelp()
		os.Exit(1)
	}

	// 确保输出目录存在
	err := os.MkdirAll(outputDir, 0o755)
	if err != nil {
		log.Fatalf("无法创建输出目录：%v", err)
	}

	// 检查输入目录中的文件类型
	shapefiles, _ := filepath.Glob(filepath.Join(inputDir, "*.shp"))
	geojsonFiles, _ := filepath.Glob(filepath.Join(inputDir, "*.geojson"))

	if len(shapefiles) > 0 {
		fmt.Printf("找到 %d 个 Shapefile，开始转换为 GeoJSON...\n", len(shapefiles))
		err := shp.BatchConvertShapefilesToGeoJSON(inputDir, outputDir)
		if err != nil {
			log.Fatalf("批量转换 Shapefile 失败：%v", err)
		}
	}

	if len(geojsonFiles) > 0 {
		fmt.Printf("找到 %d 个 GeoJSON 文件，开始转换为 Shapefile...\n", len(geojsonFiles))
		err := shp.BatchConvertGeoJSONsToShapefiles(inputDir, outputDir)
		if err != nil {
			log.Fatalf("批量转换 GeoJSON 失败：%v", err)
		}
	}

	if len(shapefiles) == 0 && len(geojsonFiles) == 0 {
		fmt.Printf("在目录 %s 中没有找到 .shp 或 .geojson 文件\n", inputDir)
	}

	fmt.Println("批量转换完成")
}
