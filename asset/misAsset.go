package asset

import (
	"AssetPathEditor/types"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MakeMisAssetMap goes through the xml files in the temp folder and reads the asset
// dependancies listed in each file, returning a map of [Asset]Asset
func MakeMisAssetMap() (AssetAssetMap, error) {
	fmt.Printf("Processing xml files")
	dotCounter := types.NewDotCounter()
	misAssetMap := make(AssetAssetMap)
	err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && (strings.Contains(path, "Network") || strings.Contains(path, "Scenery")) {
			err = misAssetMap.getFileAssets(path)
			if err != nil {
				return err
			}
			dotCounter.PrintDot()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n")

	return misAssetMap, nil
}

// getFileAssets is a function that is used by .MakeMisAssetMap to get the assets
// from a specific file
func (misAssets AssetAssetMap) getFileAssets(path string) error { // Open xml file
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
	matches := groupRegex.FindAllSubmatch(xmlBytes, -1)

	// Add s to asset map
	for i, _ := range matches {
		// Route calls for .xml scenery, but stored in Assets as .bin
		// Strange issue where &'s are represended as &amp;
		tempFilepath := strings.ReplaceAll(string(matches[i][3]), "xml", "bin")
		tempFilepath = strings.ReplaceAll(tempFilepath, `&amp;`, `&`)
		tempAsset := types.Asset{
			Provider: string(matches[i][1]),
			Product:  string(matches[i][2]),
			Filepath: tempFilepath,
		}
		misAssets[tempAsset] = types.EmptyAsset
	}
	return nil
}

// Index finds all the assets in the asset folder that are located in the provider folders
func (misAssets AssetAssetMap) Index() (AssetBoolMap, error) {
	fmt.Println("Indexing assets in Asset folder")
	dotCounter := types.NewDotCounter()
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
			misAssets.getZipAssets(path, allAssets)
		}
		dotCounter.PrintDot()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allAssets, nil
}

// getZipAssets is a function calle by Index to unzip and retrieve assets in ".ap" files
func (misAssets AssetAssetMap) getZipAssets(filename string, allAssets AssetBoolMap) error {
	filenameSlice := strings.SplitN(filename, `\`, 4)
	var buf bytes.Buffer
	cmd := exec.Command("7z.exe", "l", filename, "-ba")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return err
	}
	bufString := buf.String()
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

func (misAssets AssetAssetMap) Check(allAssets AssetBoolMap, providers ProviderMap) {
	fmt.Println("Checkng assets")
	logFile, _ := os.Create("assetlog.txt")
	rightPlace := 0
	differentPlace := 0
	notFound := 0
	capProblem := 0
	folderOrder := 0
	initialAssetCount := len(misAssets)
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
				tempAsset := types.Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				fmt.Fprintf(logFile, "CAP %v\t%v\n", misAsset, locAsset)
				capProblem++
				continue OUTER2
			}
		}
	}

	// Next check for the odd case of folder name rearraging
OUTER3:
	for misAsset, _ := range misAssets {
		for locAsset, _ := range allAssets {
			if misAssets[misAsset] != types.EmptyAsset {
				continue OUTER3
			}
			misPathSlice := strings.Split(misAsset.Filepath, `\`)
			locPathSlice := strings.Split(locAsset.Filepath, `\`)
			misBinName := misPathSlice[len(misPathSlice)-1]
			locBinName := locPathSlice[len(locPathSlice)-1]
			if strings.EqualFold(misBinName, locBinName) && strings.Contains(misAsset.Filepath, locAsset.Product) {
				tempAsset := types.Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				fmt.Fprintf(logFile, "REA %v\t%v\n", misAsset, locAsset)
				folderOrder++
				continue OUTER3
			}
		}
	}

	// Next check for bins in the provider folders
OUTER4:
	for misAsset, value := range misAssets {
		for locAsset, _ := range allAssets {
			if misAssets[misAsset] != types.EmptyAsset {
				continue OUTER4
			}
			misPathSlice := strings.Split(misAsset.Filepath, `\`)
			locPathSlice := strings.Split(locAsset.Filepath, `\`)
			misBinName := misPathSlice[len(misPathSlice)-1]
			locBinName := locPathSlice[len(locPathSlice)-1]
			locProvider, _ := providers[locAsset.Product]
			if strings.EqualFold(misBinName, locBinName) &&
				strings.EqualFold(locProvider, locAsset.Provider) { // Is this asset in one of the providers?
				tempAsset := types.Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				fmt.Fprintf(logFile, "MOV %v\t%v\n", misAsset, locAsset)
				differentPlace++
				continue OUTER4
			}
		}
		if value == types.EmptyAsset {
			fmt.Fprintf(logFile, "LST %v\n", misAsset)
			notFound++
		}
	}
	fmt.Printf("There are initially %d required assets\n", initialAssetCount)
	fmt.Printf("%v assets cannot be found\n", notFound)
	fmt.Printf("%v assets have been found but with cap errors\n", capProblem)
	fmt.Printf("%v assets have been found in the correct folder\n", rightPlace)
	fmt.Printf("%v assets have been found by rearraging folders\n", folderOrder)
	fmt.Printf("%v assets have been found, but not in the requested location\n", differentPlace)
}

// ReplaceXmlText works through a folder of xml files, and using the list of missing
// and located assets provided by missingAssetMap, substitutes the missing assets with
// the found ones.  Returns map of files that have been updated
func (misAssets AssetAssetMap) ReplaceXmlText() error {
	// string used by regex to pull groups out of the xml file
	fmt.Printf("Updating XML files")
	dotCounter := types.NewDotCounter()
	changedFiles := make(map[string]bool)
	var updatedFileBytes []byte

	err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true {
			xmlFile, err := os.OpenFile(path, os.O_RDWR, 0755)
			if err != nil {
				return err
			}
			fileBytes := make([]byte, info.Size()+4000)
			_, err = xmlFile.Read(fileBytes)
			if err != nil {
				return err
			}
			updatedFileBytes = fileBytes
			for oldAsset, newAsset := range misAssets {
				if newAsset == types.EmptyAsset {
					continue
				}
				fixNewPath := strings.ReplaceAll(newAsset.Filepath, ".bin", ".xml") // new asset is converted to match
				fixPath := strings.ReplaceAll(oldAsset.Filepath, `\`, `\\`)         // prevents \ being read as escape characters
				fixPath = strings.ReplaceAll(fixPath, ".bin", ".xml")               // In the xml doc, assets have xml extensions
				// findString is the string used by regex to find the asset lisiting in the xml (for a specific asset)
				var findString string = `<Provider d:type="cDeltaString">` + oldAsset.Provider + `</Provider>\s*` +
					`\s*<Product d:type="cDeltaString">` + oldAsset.Product + `</Product>\s*` +
					`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
					`\s*</BlueprintSetID>\s*` +
					`\s*<BlueprintID d:type="cDeltaString">` + fixPath + `</BlueprintID>`
				regex := regexp.MustCompile(findString)
				retReg := regex.Find(fileBytes) // Is the pattern located in the file?
				if retReg == nil {
					//No?
					continue
				} else {
					changedFiles[path] = true
				}
				matches := groupRegex.FindSubmatch(retReg) // put them in a slice
				if len(matches) == 0 {
					log.Fatal("There was an error parsing the groups in the regex")
				}

				// Replace the xml (now byte slice) matches
				retRegNew := bytes.Replace(retReg, matches[1], []byte(newAsset.Provider), 1)
				retRegNew = bytes.Replace(retRegNew, matches[2], []byte(newAsset.Product), 1)
				retRegNew = bytes.Replace(retRegNew, matches[3], []byte(fixNewPath), 1)
				updatedFileBytes = regex.ReplaceAll(updatedFileBytes, retRegNew)

			}
			fmt.Printf(".")
			err = xmlFile.Truncate(0)
			if err != nil {
				return err
			}
			_, err = xmlFile.WriteAt(updatedFileBytes, 0)
			if err != nil {
				return err
			}
			xmlFile.Close()
			dotCounter.PrintDot()

		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	file, err := os.Create("changed_files.txt")
	for cfile, _ := range changedFiles {
		fmt.Fprintf(file, "%v\n", cfile)
	}
	if err != nil {
		return err
	}
	return nil
}

// GetProviders lists the unique products and providers that a route uses
func (misAssets AssetAssetMap) GetProviders() map[string]string {
	fmt.Println("Getting providers for route")
	dotCounter := types.NewDotCounter()
	uniqueAssets := make(map[string]string)
	for asset, _ := range misAssets {
		if _, ok := uniqueAssets[asset.Product]; ok == false {
			uniqueAssets[asset.Product] = asset.Provider
		}
		dotCounter.PrintDot()
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
	fmt.Printf("\n\nRoute Dependancies:\n")
	fmt.Printf("%-19v%-19v\n", "Provider", "Product")
	fmt.Println("--------------------------------------")
	for _, ProvProd := range assetList {
		fmt.Printf("%-19v%-19v\n", ProvProd[0], ProvProd[1])
	}
	fmt.Printf("\n\n")
	return uniqueAssets
}
