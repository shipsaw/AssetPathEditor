package bin

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"trainTest/asset"
)

type RecordSet struct {
	XMLName xml.Name `xml:"cRecordSet"`
	Record  Record   `xml:"Record"`
}

type Record struct {
	Entities []Entity `xml:"cDynamicEntity"`
}

type Entity struct {
	BlueprintID BlueprintID `xml:"BlueprintID"`
}

type BlueprintID struct {
	AbsBlueprint AbsBlueprint `xml:"iBlueprintLibrary-cAbsoluteBlueprintID"`
}

type AbsBlueprint struct {
	BlueprintID  string       `xml:"BlueprintID"`
	BlueprintSet BlueprintSet `xml:"BlueprintSetID"`
}

type BlueprintSet struct {
	BlueprintLibSet BlueprintLibSet `xml:"iBlueprintLibrary-cBlueprintSetID"`
}

type BlueprintLibSet struct {
	Provider string `xml:"Provider"`
	Product  string `xml:"Product"`
}

func ListReqAssets(binFolder string, misAssetMap map[asset.Asset]bool) error {
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

func getFileAssets(path string, assetMap map[asset.Asset]bool) {
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
	entityCount := len(xmlStruct.Record.Entities)
	fmt.Println("Entity count = ", entityCount)
	for i := 0; i < entityCount; i++ {
		asset := asset.Asset{
			Filepath: xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintID,
			Product:  xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Product,
			Provider: xmlStruct.Record.Entities[i].BlueprintID.AbsBlueprint.BlueprintSet.BlueprintLibSet.Provider,
		}
		if assetMap[asset] == false {
			assetMap[asset] = true
		}
	}
}

func MoveXmlFiles(binFolder string, xmlFolder string) {
	err := filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
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
}
