package asset

import (
	"bytes"
	"fmt"
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

func Check(misAssets MisAssetMap, allAssets allAssetMap) {
	fmt.Printf("There are initially %d required assets\n", len(misAssets))
	rightPlace := 0
	differentPlace := 0
	notFound := 0
OUTER:
	for misAsset, value := range misAssets {
		for locAsset, _ := range allAssets {
			_, ok := allAssets[misAsset]
			if ok == true {
				rightPlace++
				delete(misAssets, misAsset)
				continue OUTER
			} else if misAsset.Filepath == locAsset.Filepath {
				tempAsset := Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				differentPlace++
				continue OUTER
			}
		}
		if value == EmptyAsset {
			fmt.Println(misAsset)
			notFound++
		}
	}
	fmt.Printf("%v assets cannot be found\n", notFound)
	fmt.Printf("%v assets have been found in the correct folder\n", rightPlace)
	fmt.Printf("%v assets have been found, but not in the requested location\n", differentPlace)
}

func Index(misAssets MisAssetMap) (allAssetMap, error) {
	allAssets := make(allAssetMap)
	err := filepath.Walk(`.\Assets`, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".bin" {
			// Seperate to find providers, products, and paths
			pathSlice := strings.SplitN(path, `\`, 4)
			if len(pathSlice) == 4 {
				asset := Asset{
					Product:  pathSlice[2],
					Provider: pathSlice[1],
					Filepath: pathSlice[3],
				}
				allAssets[asset] = false
			}
		} else if filepath.Ext(path) == ".ap" {
			GetZipAssets(path, misAssets, allAssets)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allAssets, nil
}

func GetZipAssets(filename string, misAssets MisAssetMap, allAssets allAssetMap) error {
	filenameSlice := strings.SplitN(filename, `\`, 4)
	var buf bytes.Buffer
	cmd := exec.Command("7z.exe", "l", filename, "-ba")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return err
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
	return nil
}
