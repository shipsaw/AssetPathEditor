package main

import (
	"log"
	"trainTest/asset"
	"trainTest/bin"
)

const (
	binFolder string = `.\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\Scenery\`
	xmlFolder string = `filesXml\`
)

func main() {
	misAssets, err := bin.ListReqAssets(binFolder)
	if err != nil {
		log.Fatal(err)
	}
	// Move xml file
	bin.MoveXmlFiles(binFolder, xmlFolder)
	allAssets := asset.Index(misAssets)
	asset.Check(misAssets, allAssets)
	bin.ReplaceXmlText(xmlFolder, misAssets)
	bin.MoveXmlFiles(xmlFolder, binFolder)
	bin.SerzConvert(binFolder, ".xml")
}
