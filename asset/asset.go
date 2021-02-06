package asset

/*
	asset contains the functions that are involved in processing and updating the assets listed
	in the route's xml files
*/

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"trainTest/bin"
	"trainTest/types"
)

const (
	routeFolder   string = `.\Content\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\`
	backupFolder  string = `AssetBackup\`
	replaceRoute  string = `.\Content\Routes\3a99321a-0bb2-47be-bcad-b20cfe48a945\`
	xmlFolder     string = `tempFiles\`
	sceneryFolder string = `Scenery\`
)

type AssetAssetMap map[types.Asset]types.Asset
type AssetBoolMap map[types.Asset]bool
type ProviderMap map[string]string

// GetProviders lists the unique products and providers that a route uses
func GetProviders(misAssets AssetAssetMap) map[string]string {
	uniqueAssets := make(map[string]string)
	for asset, _ := range misAssets {
		if _, ok := uniqueAssets[asset.Product]; ok == false {
			uniqueAssets[asset.Product] = asset.Provider
		}
	}

	assetList := make([][2]string, len(uniqueAssets))
	i := 0
	for product, provider := range uniqueAssets {
		assetList[i][0] = provider
		assetList[i][1] = product
		i++
	}
	sort.Slice(assetList, func(i, j int) bool {
		return assetList[i][0] > assetList[j][0]
	})
	/*
		fmt.Printf("\n\nRoute Dependancies:\n")
		fmt.Printf("%-19v%-19v\n", "Provider", "Product")
		fmt.Println("--------------------------------------")
		for _, ProvProd := range assetList {
			fmt.Printf("%-19v%-19v\n", ProvProd[0], ProvProd[1])
		}
		fmt.Printf("\n\n")
	*/
	return uniqueAssets
}

// ListReqAssets goes through the xml files in the temp folder and reads the asset
// dependancies listed in each file, returning a map of [Asset]Asset
func ListReqAssets() (AssetAssetMap, error) {
	fmt.Printf("Processing xml files")
	misAssetMap := make(AssetAssetMap)
	err := filepath.Walk(xmlFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true {
			err = getFileAssets(path, misAssetMap)
			if err != nil {
				return err
			}
		}
		fmt.Printf(".")
		return nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n")
	return misAssetMap, nil
}

//if providers[asset.Product] == asset.Provider { // If this .bin is in our providers map

// Index finds all the assets in the asset folder that are located in the provider folders
func Index(misAssets AssetAssetMap) (AssetBoolMap, error) {
	allAssets := make(AssetBoolMap)
	err := filepath.Walk(`.\Assets`, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".bin" {
			// Seperate to find providers, products, and paths
			pathSlice := strings.SplitN(path, `\`, 4)
			if len(pathSlice) == 4 { //Catches some misplaced .bins creators have placed
				asset := types.Asset{
					Product:  pathSlice[2],
					Provider: pathSlice[1],
					Filepath: pathSlice[3],
				}
				allAssets[asset] = false
			}
		} else if filepath.Ext(path) == ".ap" {
			getZipAssets(path, misAssets, allAssets)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allAssets, nil
}

// getZipAssets is a function calle by Index to unzip and retrieve assets in ".ap" files
func getZipAssets(filename string, misAssets AssetAssetMap, allAssets AssetBoolMap) error {
	fmt.Println(filename)
	filenameSlice := strings.SplitN(filename, `\`, 4)
	var buf bytes.Buffer
	cmd := exec.Command("7z.exe", "l", filename, "-ba")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return err
	}
	bufString := buf.String()
	if strings.Contains(filename, "RailsimulatorUS") {
		fmt.Println("WRITE FILE")
		ioutil.WriteFile("output.txt", []byte(bufString), 0755)
	}
	// Make a slice of all the paths in the zip
	filesSlice := strings.Split(bufString, "\n")
	for i, listing := range filesSlice {
		var tempPath string
		tempSlice := strings.Fields(listing)
		if len(tempSlice) > 5 {
			tempPath = strings.Join(tempSlice[5:], " ")
		} else if len(tempSlice) == 5 {
			tempPath = tempSlice[4]
		}
		filesSlice[i] = tempPath
	}

	for misAsset, _ := range misAssets {
		for _, path := range filesSlice {
			if strings.EqualFold(misAsset.Filepath, path) { //Match the bin name
				asset := types.Asset{
					Product:  filenameSlice[2],
					Provider: filenameSlice[1],
					Filepath: misAsset.Filepath,
				}
				allAssets[asset] = false
			}
		}
	}
	return nil
}

func Check(misAssets AssetAssetMap, allAssets AssetBoolMap, providers ProviderMap) {
	fmt.Printf("There are initially %d required assets\n", len(misAssets))
	rightPlace := 0
	differentPlace := 0
	notFound := 0
	capProblem := 0
	// First check for bins in the correct location
OUTER:
	for misAsset, _ := range misAssets {
		for _, _ = range allAssets {
			_, ok := allAssets[misAsset]
			if ok == true {
				rightPlace++
				delete(misAssets, misAsset)
				continue OUTER
			}
		}
	}

	// Next check for capitalization problems
OUTER2:
	for misAsset, _ := range misAssets {
		for locAsset, _ := range allAssets {
			misFullPath := misAsset.Provider + misAsset.Product + misAsset.Filepath
			locFullPath := locAsset.Provider + locAsset.Product + locAsset.Filepath
			if strings.EqualFold(misFullPath, locFullPath) && !strings.Contains(misFullPath, locFullPath) {
				fmt.Println("Cap Probem!")
				tempAsset := types.Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				capProblem++
				continue OUTER2
			}
		}
	}
	// Next check for bins in the provider folders
OUTER3:
	for misAsset, value := range misAssets {
		for locAsset, _ := range allAssets {
			if misAsset.Filepath == `scenery\vegetation\tree_misc_large_line01.bin` &&
				locAsset.Filepath == `scenery\vegetation\tree_misc_large_line01.bin` {
				fmt.Println(`FOUND scenery\vegetation\tree_misc_large_line01.bin`)
			}
			misPathSlice := strings.Split(misAsset.Filepath, `\`)
			locPathSlice := strings.Split(locAsset.Filepath, `\`)
			misBinName := misPathSlice[len(misPathSlice)-1]
			locBinName := locPathSlice[len(locPathSlice)-1]
			locProvider, _ := providers[locAsset.Product]
			if misBinName == locBinName && strings.EqualFold(locProvider, locAsset.Provider) { // Is this asset in one of the providers?
				tempAsset := types.Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				differentPlace++
				continue OUTER3
			}
		}
		fmt.Println("NOT FOUND")
		if value == types.EmptyAsset {
			fmt.Println(misAsset)
			notFound++
		}
	}
	fmt.Printf("%v assets cannot be found\n", notFound)
	fmt.Printf("%v assets have been found but with cap errors\n", capProblem)
	fmt.Printf("%v assets have been found in the correct folder\n", rightPlace)
	fmt.Printf("%v assets have been found, but not in the requested location\n", differentPlace)
}

// ReplaceXmlText works through a folder of xml files, and using the list of missing
// and located assets provided by missingAssetMap, substitutes the missing assets with
// the found ones
func ReplaceXmlText(misAssets AssetAssetMap) error {
	// string used by regex to pull groups out of the xml file
	var groupedString string = `\s*<Provider d:type="cDeltaString">(.+)</Provider>\s*` +
		`\s*<Product d:type="cDeltaString">(.+)</Product>\s*` +
		`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
		`\s*</BlueprintSetID>\s*` +
		`\s*<BlueprintID d:type="cDeltaString">(.+)</BlueprintID>`
	fmt.Printf("Updating XML files")

	err := filepath.Walk(xmlFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true {
			xmlFile, err := os.OpenFile(path, os.O_RDWR, 0755)
			if err != nil {
				return err
			}
			fileBytes := make([]byte, info.Size())
			_, err = xmlFile.Read(fileBytes)
			if err != nil {
				return err
			}

			for oldAsset, newAsset := range misAssets {
				if newAsset == types.EmptyAsset {
					continue
				}
				fixPath := strings.ReplaceAll(oldAsset.Filepath, `\`, `\\`) // prevents \ being read as escape characters
				fixPath = strings.ReplaceAll(fixPath, ".bin", ".xml")       // In the xml doc, assets have xml extensions
				// findString is the string used by regex to find the asset lisiting in the xml (for a specific asset)
				var findString string = `<Provider d:type="cDeltaString">` + oldAsset.Provider + `</Provider>\s*` +
					`\s*<Product d:type="cDeltaString">` + oldAsset.Product + `</Product>\s*` +
					`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
					`\s*</BlueprintSetID>\s*` +
					`\s*<BlueprintID d:type="cDeltaString">` + fixPath + `</BlueprintID>`
				regex := regexp.MustCompile(findString)
				retReg := regex.Find(fileBytes) // Is the pattern located in the file?
				if retReg == nil {
					continue
				}
				groupRegex := regexp.MustCompile(groupedString) // Pull the groups from the match
				matches := groupRegex.FindSubmatch(retReg)      // put them in a slice
				if len(matches) == 0 {
					log.Fatal("There was an error parsing the groups in the regex")
				}

				fixNewPath := strings.ReplaceAll(newAsset.Filepath, ".bin", ".xml") // new asset is converted to match
				// Replace the xml (now byte slice) matches
				retRegNew := bytes.Replace(retReg, matches[1], []byte(newAsset.Provider), 1)
				retRegNew = bytes.Replace(retRegNew, matches[2], []byte(newAsset.Product), 1)
				retRegNew = bytes.Replace(retRegNew, matches[3], []byte(fixNewPath), 1)
				fileBytes = bytes.Replace(fileBytes, retReg, retRegNew, -1)

			}
			fmt.Printf(".")
			err = xmlFile.Truncate(0) // make sure we overwrite the old xml doc
			if err != nil {
				return err
			}
			_, err = xmlFile.WriteAt(fileBytes, 0)
			if err != nil {
				return err
			}
			xmlFile.Close()
		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}

// getFileAssets is a function that is used by .ListReqAssets to get the assets
// from a specific file
func getFileAssets(path string, misAssets AssetAssetMap) error { // Open xml file
	xmlFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer xmlFile.Close()
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	xmlBytes := make([]byte, info.Size())
	numRead, err := xmlFile.Read(xmlBytes)
	if err != nil {
		return err
	}
	// Check that xml processing looks correct
	if numRead != int(info.Size()) {
		log.Fatal("Size mismatch on file ", path)
	}
	// Unmarshal xml
	xmlStruct := bin.RecordSet{}
	err = xml.Unmarshal(xmlBytes, &xmlStruct)
	if err != nil {
		return err
	}
	// Add s to asset map
	entityCount := len(xmlStruct.Record.Entities)
	for i := 0; i < entityCount; i++ {
		// Route calls for .xml scenery, but stored in Assets as .bin
		tempFilepath := xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintID
		tempFilepath = strings.ReplaceAll(tempFilepath, "xml", "bin")
		tempAsset := types.Asset{
			Filepath: tempFilepath,
			Product:  xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Product,
			Provider: xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Provider,
		}
		misAssets[tempAsset] = types.EmptyAsset
	}
	return nil
}

// ListProviders list the products and providers used by a route.  It does the normal setup/teardown in the route
// but doesn't care about backups because nothing is changed in the bin files
func ListProviders(route string) (map[string]string, error) {
	routeBackup := route + backupFolder
	err := bin.Setup(route, routeBackup)
	if err != nil {
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}
	misAssets, err := ListReqAssets()
	if err != nil {
		bin.Teardown(routeBackup, false)
		log.Fatal(err)
	}
	providers := GetProviders(misAssets)
	bin.Revert(route, routeBackup)
	bin.Teardown(routeBackup, true)
	return providers, nil
}

func UpdateRoute(route string, providers ProviderMap) error {
	routeBackup := route + backupFolder
	err := bin.Setup(route, routeBackup)
	if err != nil {
		bin.Teardown(routeBackup, true)
		log.Fatal(err)
	}
	misAssets, err := ListReqAssets()
	if err != nil {
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, true)
	}

	locAssets, err := Index(misAssets)
	if err != nil {
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, true)
	}
	Check(misAssets, locAssets, providers)
	err = ReplaceXmlText(misAssets)
	if err != nil {
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
	}
	err = bin.SerzConvert(".xml")
	if err != nil {
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
	}
	bin.MoveAssetFiles(xmlFolder, route+sceneryFolder, ".bin")
	if err != nil {
		bin.Revert(route, routeBackup)
		bin.Teardown(routeBackup, false)
	}
	bin.Teardown(routeBackup, false)

	return nil
}
