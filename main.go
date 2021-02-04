package main

import (
	"fmt"
	"log"
	"trainTest/asset"
	"trainTest/bin"
)

const (
	routeFolder  string = `.\Content\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\`
	backupFolder string = `AssetBackup\`
	replaceRoute string = `.\Content\Routes\3a99321a-0bb2-47be-bcad-b20cfe48a945\`
)

func main() {
	if err := ListProviders(replaceRoute); err != nil {
		log.Fatal(err)
	}
	/*
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
		_ = asset.GetProviders(misAssets)
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
		bin.SerzConvert(routeFolder, ".xml")
		if err != nil {
			bin.Teardown(backupFolder, false)
			log.Fatal(err)
		}
		fmt.Println("press key to revert")
		fmt.Scanln()
		bin.Revert(routeFolder, backupFolder)
		bin.Teardown(backupFolder, false)
	*/
}

// ListProviders list the products and providers used by a route.  It does the normal setup/teardown in the route
// but doesn't care about backups because nothing is changed in the bin files
func ListProviders(route string) error {
	routeBackup := route + backupFolder
	fmt.Println("Running setup")
	err := bin.Setup(route, routeBackup)
	if err != nil {
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}
	misAssets, err := bin.ListReqAssets()
	if err != nil {
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	_ = asset.GetProviders(misAssets)
	bin.Revert(route, routeBackup)
	bin.Teardown(routeBackup, true)
	return nil
}
