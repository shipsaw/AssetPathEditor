package asset

/*
	asset contains the functions that are involved in processing and updating the assets listed
	in the route's xml files
*/

import (
	"AssetPathEditor/bin"
	"AssetPathEditor/types"
	"fmt"
	"log"
	"regexp"
)

const (
	backupFolder string = `AssetBackup\`
	workspace    string = `tempFolder\`
)

var groupedString string = `\s*<Provider d:type="cDeltaString">(.+)</Provider>\s*` +
	`\s*<Product d:type="cDeltaString">(.+)</Product>\s*` +
	`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
	`\s*</BlueprintSetID>\s*` +
	`\s*<BlueprintID d:type="cDeltaString">(.+)</BlueprintID>`

var groupRegex *regexp.Regexp = regexp.MustCompile(groupedString) // Pull the groups from the match

type AssetAssetMap map[types.Asset]types.Asset
type AssetBoolMap map[types.Asset]bool
type ProviderMap map[string]string

// ListProviders list the products and providers used by a route.  It does the normal setup/teardown in the route
// but doesn't care about backups because nothing is changed in the bin files
func ListProviders(route string) (map[string]string, error) {
	routeBackup := route + backupFolder
	err := bin.Setup(route)
	if err != nil {
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}
	err = bin.SerzConvert(".bin")
	if err != nil {
		return nil, err
	}
	misAssets, err := MakeMisAssetMap()
	if err != nil {
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	providers := misAssets.GetProviders()
	bin.Revert(route, routeBackup)
	bin.Teardown(routeBackup, true)
	return providers, nil
}

func UpdateRoute(route string, providers ProviderMap) error {
	routeBackup := route + backupFolder
	err := bin.Setup(route)
	if err != nil {
		fmt.Println("SETUP ERROR")
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}

	err = bin.SerzConvert(".bin")
	if err != nil {
		fmt.Println("SERZ ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}

	misAssets, err := MakeMisAssetMap()
	if err != nil {
		fmt.Println("LIST REQ ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}

	locAssets, err := misAssets.Index()
	if err != nil {
		fmt.Println("INDEX ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}
	misAssets.Check(locAssets, providers)
	err = misAssets.ReplaceXmlText()
	if err != nil {
		fmt.Println("INDEX ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	err = bin.SerzConvert(".xml")
	if err != nil {
		fmt.Println("SECOND SERZ ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	bin.MoveAssetFiles(workspace, route, ".bin")
	if err != nil {
		fmt.Println("SECOND MOVEASSET ERROR")
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	bin.Teardown(routeBackup, false)

	return nil
}
