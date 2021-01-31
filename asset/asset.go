package asset

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Asset struct {
	Product  string
	Provider string
	Filepath string
}

func Print(misAssetMap map[Asset]bool) {
	assetList := make([]Asset, len(misAssetMap))
	i := 0
	for asset, _ := range misAssetMap {
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

func Check(misAssetMap map[Asset]bool) {
	i := 0
	for asset, _ := range misAssetMap {
		assetPath := `.\Assets\` + asset.Provider + `\` + asset.Product + `\` + asset.Filepath
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			//fmt.Println("File is missing: ", assetPath)
			i++
		} else if err == nil {
			misAssetMap[asset] = false
		} else {
			log.Fatal(err)
		}
	}
	fmt.Printf("There are %v assets missing\n", i)
}

func Find(misAssetMap, allAssetMap map[Asset]bool) {
	// Create map
	for misAsset, missing := range misAssetMap {
		for allAsset, _ := range allAssetMap {
			if misAsset.Filepath == allAsset.Filepath && missing == true {
				fmt.Printf("Asset called for in %v/%v located in %v/%v\n", misAsset.Provider, misAsset.Product, allAsset.Provider, allAsset.Product)
			}
		}
	}
}

func Index() map[Asset]bool {
	allAssetMap := make(map[Asset]bool)
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
			allAssetMap[asset] = true
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return allAssetMap
}
