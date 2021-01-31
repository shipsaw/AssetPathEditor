package main

import (
	"fmt"
	"log"
	"sort"
	"trainTest/asset"
	"trainTest/bin"
)

const (
	binFolder string = `.\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\Scenery\`
	xmlFolder string = "filesXml" + `\`
)

func main() {
	misAssetMap := make(map[asset.Asset]bool)
	err := bin.ListReqAssets(binFolder, misAssetMap)
	if err != nil {
		log.Fatal(err)
	}
	// Move xml file
	fmt.Println("Moving xml files")
	bin.MoveXmlFiles(binFolder, xmlFolder)
	fmt.Println("Completed move")

	assetList := make([]asset.Asset, len(misAssetMap))
	fmt.Println(len(misAssetMap))
	i := 0
	for asset, _ := range misAssetMap {
		assetList[i] = asset
		i++
	}
	sort.Slice(assetList, func(i, j int) bool {
		return assetList[i].Product > assetList[j].Product
	})
	for _, asset := range assetList {
		fmt.Printf("Prod: %-19vProv: %-19vPath: %v\n", asset.Product, asset.Provider, asset.Filepath)
	}
}
