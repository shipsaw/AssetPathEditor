package bin

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"trainTest/asset"
)

func SerzConvert(folder, ext string) {
	fmt.Printf("Converting files")
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() != true && filepath.Ext(path) == ext {
			cmd := exec.Command("serz.exe", path)
			if err := cmd.Run(); err != nil {
				fmt.Println("Error: ", err)
			}
			err := os.Remove(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf(".")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")
}
func ListReqAssets(binFolder string) (asset.MisAssetMap, error) {
	SerzConvert(binFolder, ".bin")
	fmt.Printf("Processing xml files")
	misAssetMap := make(asset.MisAssetMap)
	err := filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() != true && filepath.Ext(path) == ".xml" {
			getFileAssets(path, misAssetMap)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")
	return misAssetMap, nil
}

func getFileAssets(path string, misAssets asset.MisAssetMap) {
	fmt.Printf(".")
	xmlStruct := RecordSet{}
	// Open xml file
	xmlFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()
	info, err := os.Lstat(path)
	if err != nil {
		log.Fatal(err)
	}
	xmlBytes := make([]byte, info.Size())
	numRead, err := xmlFile.Read(xmlBytes)
	if err != nil {
		log.Fatal(err)
	}
	// Check that xml processing looks correct
	if numRead != int(info.Size()) {
		log.Fatal("Size mismatch on file ", path)
	}
	// Unmarshal xml
	err = xml.Unmarshal(xmlBytes, &xmlStruct)
	if err != nil {
		log.Fatal(err)
	}
	// Add assets to asset map
	entityCount := len(xmlStruct.Record.Entities)
	for i := 0; i < entityCount; i++ {
		// Route calls for .xml scenery, but stored in Assets as .bin
		tempFilepath := xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintID
		tempFilepath = strings.ReplaceAll(tempFilepath, "xml", "bin")
		tempAsset := asset.Asset{
			Filepath: tempFilepath,
			Product:  xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Product,
			Provider: xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Provider,
		}
		misAssets[tempAsset] = asset.EmptyAsset
	}
}

func MoveXmlFiles(oldLoc string, newLoc string) {
	err := filepath.Walk(oldLoc, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() != true && filepath.Ext(path) == ".xml" {
			newPath := newLoc + filepath.Base(path)
			err = os.Rename(path, newPath)
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ReplaceXmlText(xmlFolder string, misAssets asset.MisAssetMap) {

	var groupedString string = `\s*<Provider d:type="cDeltaString">(.+)</Provider>\s*` +
		`\s*<Product d:type="cDeltaString">(.+)</Product>\s*` +
		`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
		`\s*</BlueprintSetID>\s*` +
		`\s*<BlueprintID d:type="cDeltaString">(.+)</BlueprintID>`
	fmt.Printf("Updating XML files")
	err := filepath.Walk(xmlFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() != true {
			xmlFile, err := os.OpenFile(path, os.O_RDWR, 0755)
			if err != nil {
				log.Fatal(err)
			}
			defer xmlFile.Close()
			fileBytes := make([]byte, info.Size())
			_, err = xmlFile.Read(fileBytes)
			if err != nil {
				log.Fatal(err)
			}

			for oldAsset, newAsset := range misAssets {
				if newAsset == asset.EmptyAsset {
					continue
				}
				fixPath := strings.ReplaceAll(oldAsset.Filepath, `\`, `\\`)
				fixPath = strings.ReplaceAll(fixPath, ".bin", ".xml")
				var findString string = `<Provider d:type="cDeltaString">` + oldAsset.Provider + `</Provider>\s*` +
					`\s*<Product d:type="cDeltaString">` + oldAsset.Product + `</Product>\s*` +
					`\s*</iBlueprintLibrary-cBlueprintSetID>\s*` +
					`\s*</BlueprintSetID>\s*` +
					`\s*<BlueprintID d:type="cDeltaString">` + fixPath + `</BlueprintID>`
				regex := regexp.MustCompile(findString)
				retReg := regex.Find(fileBytes)
				if retReg == nil {
					continue
				}
				groupRegex := regexp.MustCompile(groupedString)
				matches := groupRegex.FindSubmatch(retReg)
				if len(matches) == 0 {
					log.Fatal("There was an error parsing the groups in the regex")
				}

				retRegNew := bytes.Replace(retReg, matches[1], []byte(newAsset.Provider), 1)
				retRegNew = bytes.Replace(retRegNew, matches[2], []byte(newAsset.Product), 1)
				retRegNew = bytes.Replace(retRegNew, matches[3], []byte(newAsset.Filepath), 1)
				fileBytes = bytes.Replace(fileBytes, retReg, retRegNew, -1)
				if err != nil {
					log.Fatal(err)
				}
				xmlFile.Write(fileBytes)
			}
			fmt.Printf(".")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")
}
