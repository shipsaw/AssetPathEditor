package asset

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Asset struct {
	Product  string
	Provider string
	Filepath string
}

var EmptyAsset = Asset{
	Product:  "",
	Provider: "",
	Filepath: "",
}

type MisAssetMap map[Asset]Asset
type allAssetMap map[Asset]bool

func Print(misAssets MisAssetMap) {
	/*assetList := make([]Asset, len(misAssets))
	i := 0
	for misAsset, foundAsset := range misAssets {
		assetList[i] = asset
		i++
	}
	sort.Slice(assetList, func(i, j int) bool {
		return assetList[i].Product > assetList[j].Product
	})
	/*
		for _, asset := range assetList {
			fmt.Printf("Prod: %-19vProv: %-19vPath: %v\n", asset.Product, asset.Provider, asset.Filepath)
		}
	*/
}

func Check(misAssets MisAssetMap) {
	i := 0
	for asset, _ := range misAssets {
		assetPath := `.\Assets\` + asset.Provider + `\` + asset.Product + `\` + asset.Filepath
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			//fmt.Println("File is missing: ", assetPath)
			i++
		} else if err == nil {
			delete(misAssets, asset)
		} else {
			log.Fatal(err)
		}
	}
	fmt.Printf("Initially, there are %v assets missing\n", i)
}

func Find(misAssets MisAssetMap, allAssets allAssetMap) {
	// Create map
	i := 0
	for misAsset, _ := range misAssets {
		for allAsset, _ := range allAssets {
			if misAsset.Filepath == allAsset.Filepath {
				tempAsset := Asset{
					Product:  allAsset.Product,
					Provider: allAsset.Provider,
					Filepath: allAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				i++
			}
		}
	}
	fmt.Printf("Found %v assets in other Asset folders\n", i)
}

func Index(misAssets MisAssetMap) allAssetMap {
	allAssets := make(allAssetMap)
	err := filepath.Walk(`.\Assets`, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if filepath.Ext(path) == ".bin" {
			// Seperate to find providers, products, and paths
			pathSlice := strings.SplitN(path, `\`, 4)
			asset := Asset{
				Product:  pathSlice[2],
				Provider: pathSlice[1],
				Filepath: pathSlice[3],
			}
			allAssets[asset] = true
		} else if filepath.Ext(path) == ".ap" {
			GetZipAssets(path, misAssets, allAssets)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return allAssets
}

func GetZipAssets(filename string, misAssets MisAssetMap, allAssets allAssetMap) {
	filenameSlice := strings.SplitN(filename, `\`, 4)
	var buf bytes.Buffer
	cmd := exec.Command("7z.exe", "l", filename, "-ba")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	zipString := buf.String()
	for misAsset, _ := range misAssets {
		if strings.Contains(zipString, misAsset.Filepath) {
			asset := Asset{
				Product:  filenameSlice[2],
				Provider: filenameSlice[1],
				Filepath: misAsset.Filepath,
			}
			allAssets[asset] = true
		}
	}
}
