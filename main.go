package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var file string

const (
	binFolder string = ".\\" + "filesBin"
	xmlFolder string = "filesXml" + "\\"
)

type RecordSet struct {
	XMLName xml.Name `xml:"cRecordSet"`
	Records Record   `xml:"Record"`
}

type Record struct {
	Entities []Entity `xml:"cDynamicEntity"`
}

type Entity struct {
	Blueprints BlueprintID `xml:"BlueprintID"`
}

type BlueprintID struct {
	BlueprintLib AbsBluePrint `xml:"iBlueprintLibrary-cAbsoluteBlueprintID"`
}

type AbsBluePrint struct {
	BlueprintID string `xml:"BlueprintID"`
}

func main() {
	misAssetMap := make(map[string]bool)
	err := listReqAssets(binFolder, misAssetMap)
	// Move xml file
	err = filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() != true && filepath.Ext(path) == ".xml" {
			splitPath := strings.SplitAfter(path, "\\")
			newPath := xmlFolder + splitPath[1]
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
	assetList := make([]string, len(misAssetMap))
	i := 0
	for asset, _ := range misAssetMap {
		assetList[i] = asset
		i++
	}
	sort.Strings(assetList)
	/*for _, asset := range assetList {
		fmt.Println(asset)
	}
	*/
}

func listReqAssets(binFolder string, misAssetMap map[string]bool) error {
	err := filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && filepath.Ext(path) == ".bin" {
			getFileAssets(path, misAssetMap)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func getFileAssets(path string, assetMap map[string]bool) {
	fmt.Println("Processing file ", path)
	xmlStruct := RecordSet{}
	pathXml := strings.ReplaceAll(path, "bin", "xml")
	//Run serz on file
	cmd := exec.Command("serz.exe", path)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
	// Open xml file
	xmlFile, err := os.Open(pathXml)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()
	info, err := os.Lstat(pathXml)
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
	entityCount := len(xmlStruct.Records.Entities)
	for i := 0; i < entityCount; i++ {
		asset := xmlStruct.Records.Entities[i].Blueprints.BlueprintLib.BlueprintID
		if assetMap[asset] == false {
			assetMap[asset] = true
		}
	}
}
