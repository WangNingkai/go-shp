package shp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConvertShapefileToGeoJSON 将 Shapefile 转换为 GeoJSON 文件.
func ConvertShapefileToGeoJSON(shapefilePath, geojsonPath string) error {
	return ConvertShapefileToGeoJSONWithOptions(shapefilePath, geojsonPath, false)
}

// ConvertShapefileToGeoJSONWithOptions 将 Shapefile 转换为 GeoJSON 文件，支持选项.
func ConvertShapefileToGeoJSONWithOptions(shapefilePath, geojsonPath string, ignoreCorrupted bool, compact ...bool) error {
	converter := GeoJSONConverter{}

	var geoJSON *GeoJSON
	var err error

	if ignoreCorrupted {
		geoJSON, err = converter.ShapefileToGeoJSONWithOptions(shapefilePath, WithIgnoreCorruptedShapes(true))
	} else {
		geoJSON, err = converter.ShapefileToGeoJSON(shapefilePath)
	}

	if err != nil {
		return fmt.Errorf("failed to convert shapefile to GeoJSON: %v", err)
	}

	// Compact flag: optional variadic boolean; if provided and true, write compact JSON
	isCompact := len(compact) > 0 && compact[0]
	err = converter.SaveGeoJSONToFile(geoJSON, geojsonPath, isCompact)
	if err != nil {
		return fmt.Errorf("failed to save GeoJSON file: %v", err)
	}

	return nil
}

// ConvertShapefileToGeoJSONStream 以流式方式将 Shapefile 转为 GeoJSON（紧凑格式），更节省内存。
func ConvertShapefileToGeoJSONStream(shapefilePath, geojsonPath string, skipCorrupted bool) error {
	converter := GeoJSONConverter{}
	f, err := os.Create(geojsonPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if skipCorrupted {
		return converter.ShapefileToGeoJSONStream(shapefilePath, f, WithIgnoreCorruptedShapes(true))
	}
	return converter.ShapefileToGeoJSONStream(shapefilePath, f)
}

// ConvertShapefileToGeoJSONString 将 Shapefile 转换为 GeoJSON 字符串.
func ConvertShapefileToGeoJSONString(shapefilePath string) (string, error) {
	converter := GeoJSONConverter{}

	geoJSON, err := converter.ShapefileToGeoJSON(shapefilePath)
	if err != nil {
		return "", fmt.Errorf("failed to convert shapefile to GeoJSON: %v", err)
	}

	data, err := json.MarshalIndent(geoJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal GeoJSON to string: %v", err)
	}

	return string(data), nil
}

// ConvertGeoJSONToShapefile 将 GeoJSON 文件转换为 Shapefile.
func ConvertGeoJSONToShapefile(geojsonPath, shapefilePath string) error {
	converter := GeoJSONConverter{}

	geoJSON, err := converter.LoadGeoJSONFromFile(geojsonPath)
	if err != nil {
		return fmt.Errorf("failed to load GeoJSON file: %v", err)
	}

	err = converter.GeoJSONToShapefile(geoJSON, shapefilePath)
	if err != nil {
		return fmt.Errorf("failed to convert GeoJSON to shapefile: %v", err)
	}

	return nil
}

// ShapeToGeoJSONString 将单个 Shape 转换为 GeoJSON 字符串.
func ShapeToGeoJSONString(shape Shape) (string, error) {
	converter := GeoJSONConverter{}

	geometry, err := converter.ShapeToGeoJSON(shape)
	if err != nil {
		return "", err
	}

	// Create a simple GeoJSON object
	geoJSON := &GeoJSON{
		Type:       "Feature",
		Geometry:   geometry,
		Properties: make(map[string]interface{}),
	}

	data, err := json.MarshalIndent(geoJSON, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ConvertShapefileToGeoJSONSkipCorrupted 将 Shapefile 转换为 GeoJSON 文件，跳过损坏的shape.
func ConvertShapefileToGeoJSONSkipCorrupted(shapefilePath, geojsonPath string) error {
	return ConvertShapefileToGeoJSONWithOptions(shapefilePath, geojsonPath, true)
}

func BatchConvertShapefilesToGeoJSON(inputDir, outputDir string) error {
	return BatchConvertShapefilesToGeoJSONWithOptions(inputDir, outputDir, false)
}

// BatchConvertShapefilesToGeoJSONWithOptions 批量转换 Shapefile 到 GeoJSON，支持静默模式.
func BatchConvertShapefilesToGeoJSONWithOptions(inputDir, outputDir string, silent bool) error {
	// 查找所有 .shp 文件
	shapefiles, err := filepath.Glob(filepath.Join(inputDir, "*.shp"))
	if err != nil {
		return err
	}

	for _, shapefile := range shapefiles {
		basename := strings.TrimSuffix(filepath.Base(shapefile), ".shp")
		geojsonPath := filepath.Join(outputDir, basename+".geojson")

		if !silent {
			fmt.Printf("Converting %s to %s...\n", shapefile, geojsonPath)
		}

		err := ConvertShapefileToGeoJSON(shapefile, geojsonPath)
		if err != nil {
			if !silent {
				fmt.Printf("Error converting %s: %v\n", shapefile, err)
			}
			continue
		}

		if !silent {
			fmt.Printf("Successfully converted %s\n", shapefile)
		}
	}

	return nil
}

// BatchConvertGeoJSONsToShapefiles 批量转换 GeoJSON 到 Shapefile.
func BatchConvertGeoJSONsToShapefiles(inputDir, outputDir string) error {
	return BatchConvertGeoJSONsToShapefilesWithOptions(inputDir, outputDir, false)
}

// BatchConvertGeoJSONsToShapefilesWithOptions 批量转换 GeoJSON 到 Shapefile，支持静默模式.
func BatchConvertGeoJSONsToShapefilesWithOptions(inputDir, outputDir string, silent bool) error {
	// 查找所有 .geojson 文件
	geojsonFiles, err := filepath.Glob(filepath.Join(inputDir, "*.geojson"))
	if err != nil {
		return err
	}

	for _, geojsonFile := range geojsonFiles {
		basename := strings.TrimSuffix(filepath.Base(geojsonFile), ".geojson")
		shapefilePath := filepath.Join(outputDir, basename+".shp")

		if !silent {
			fmt.Printf("Converting %s to %s...\n", geojsonFile, shapefilePath)
		}

		err := ConvertGeoJSONToShapefile(geojsonFile, shapefilePath)
		if err != nil {
			if !silent {
				fmt.Printf("Error converting %s: %v\n", geojsonFile, err)
			}
			continue
		}

		if !silent {
			fmt.Printf("Successfully converted %s\n", geojsonFile)
		}
	}

	return nil
}
