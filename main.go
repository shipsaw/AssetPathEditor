package main

import (
	"fmt"
	"log"
	"trainTest/asset"
	"trainTest/bin"
)

const (
	routeFolder  string = `.\Content\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\`
	xmlFolder    string = `tempFiles\`
	backupFolder string = routeFolder + `AssetBackup\`
)

func main() {
	fmt.Println("Running setup")
	err := bin.Setup(routeFolder, backupFolder)
	if err != nil {
		bin.Teardown(backupFolder, true)
		log.Fatal(err)
	}
	misAssets, err := bin.ListReqAssets(routeFolder)
	if err != nil {
		bin.Teardown(backupFolder, false)
		log.Fatal(err)
	}
	asset.GetProviders(misAssets)
	/*
		// Move xml file
		bin.MoveXmlFiles(routeFolder, xmlFolder)
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		allAssets, err := asset.Index(misAssets)
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		if err != nil {
			log.Print(err)
		}

		asset.Check(misAssets, allAssets)
		bin.ReplaceXmlText(xmlFolder, misAssets)
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		bin.MoveXmlFiles(xmlFolder, routeFolder)
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		bin.SerzConvert(routeFolder, ".xml")
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		fmt.Println("press key to revert")
		fmt.Scanln()
		bin.Revert(routeFolder, backupFolder)
	*/
	bin.Teardown(backupFolder, false)
}
