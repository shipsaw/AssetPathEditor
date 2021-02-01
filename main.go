package main

import (
	"fmt"
	"log"
	"trainTest/asset"
	"trainTest/bin"
)

const (
	binFolder string = `.\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\Scenery\`
	xmlFolder string = "filesXml" + `\`
)

func main() {
	misAssets, err := bin.ListReqAssets(binFolder)
	if err != nil {
		log.Fatal(err)
	}
	// Move xml file
	bin.MoveXmlFiles(binFolder, xmlFolder)

	fmt.Printf("Route has %v asset requirements\n", len(misAssets))
	asset.Check(misAssets)
	allAssetMap := asset.Index(misAssets)
	asset.Find(misAssets, allAssetMap)
	i := 0
	for misAsset, foundAsset := range misAssets {
		if foundAsset != asset.EmptyAsset {
			fmt.Printf("%-10v %-18v       %-20v %-10v\n", misAsset.Provider, misAsset.Product, foundAsset.Provider, foundAsset.Product)
		} else {
			i++
		}
	}
	fmt.Printf("Process completed: There are still %v assets missing", i)

}
