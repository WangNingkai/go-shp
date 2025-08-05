package shp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ConvertShapefileToGeoJSON 将 Shapefile 转换为 GeoJSON 文件
func ConvertShapefileToGeoJSON(shapefilePath, geojsonPath string) error {
	converter := GeoJSONConverter{}

	geoJSON, err := converter.ShapefileToGeoJSON(shapefilePath)
	if err != nil {
		return fmt.Errorf("failed to convert shapefile to GeoJSON: %v", err)
	}

	err = converter.SaveGeoJSONToFile(geoJSON, geojsonPath)
	if err != nil {
		return fmt.Errorf("failed to save GeoJSON file: %v", err)
	}

	return nil
}

// ConvertGeoJSONToShapefile 将 GeoJSON 文件转换为 Shapefile
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

// ShapeToGeoJSONString 将单个 Shape 转换为 GeoJSON 字符串
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

// BatchConvertShapefilesToGeoJSON 批量转换 Shapefile 到 GeoJSON
func BatchConvertShapefilesToGeoJSON(inputDir, outputDir string) error {
	// 查找所有 .shp 文件
	shapefiles, err := filepath.Glob(filepath.Join(inputDir, "*.shp"))
	if err != nil {
		return err
	}

	for _, shapefile := range shapefiles {
		basename := strings.TrimSuffix(filepath.Base(shapefile), ".shp")
		geojsonPath := filepath.Join(outputDir, basename+".geojson")

		fmt.Printf("Converting %s to %s...\n", shapefile, geojsonPath)

		err := ConvertShapefileToGeoJSON(shapefile, geojsonPath)
		if err != nil {
			fmt.Printf("Error converting %s: %v\n", shapefile, err)
			continue
		}

		fmt.Printf("Successfully converted %s\n", shapefile)
	}

	return nil
}

// BatchConvertGeoJSONsToShapefiles 批量转换 GeoJSON 到 Shapefile
func BatchConvertGeoJSONsToShapefiles(inputDir, outputDir string) error {
	// 查找所有 .geojson 文件
	geojsonFiles, err := filepath.Glob(filepath.Join(inputDir, "*.geojson"))
	if err != nil {
		return err
	}

	for _, geojsonFile := range geojsonFiles {
		basename := strings.TrimSuffix(filepath.Base(geojsonFile), ".geojson")
		shapefilePath := filepath.Join(outputDir, basename+".shp")

		fmt.Printf("Converting %s to %s...\n", geojsonFile, shapefilePath)

		err := ConvertGeoJSONToShapefile(geojsonFile, shapefilePath)
		if err != nil {
			fmt.Printf("Error converting %s: %v\n", geojsonFile, err)
			continue
		}

		fmt.Printf("Successfully converted %s\n", geojsonFile)
	}

	return nil
}
