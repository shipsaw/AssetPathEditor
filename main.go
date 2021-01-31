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
	misAssetMap, err := bin.ListReqAssets(binFolder)
	if err != nil {
		log.Fatal(err)
	}
	// Move xml file
	bin.MoveXmlFiles(binFolder, xmlFolder)

	//asset.Print(misAssetMap)
	fmt.Printf("Route has %v asset requirements\n", len(misAssetMap))
	asset.Check(misAssetMap)
	allAssetMap := asset.Index(misAssetMap)
	asset.Find(misAssetMap, allAssetMap)
	i := 0
	for _, missing := range misAssetMap {
		if missing == true {
			i++
		}
	}
	fmt.Println("Final missing assets: ", i)
}
