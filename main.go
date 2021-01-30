package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const testFile string = "+000000+000000"

var serzExe = exec.Cmd{
	Path:   "./files/serz.exe",
	Args:   []string{"./serz.exe", "./files/" + testFile + ".bin"},
	Stdout: os.Stdout,
}

type RecordSet struct {
	XMLName xml.Name `xml:"cRecordSet"`
	Records []Record `xml:"Record"`
}

type Record struct {
	XMLName  xml.Name `xml:"Record"`
	Entities []Entity `xml:"cDynamicEntity"`
}

type Entity struct {
	XMLName    xml.Name      `xml:"cDynamicEntity"`
	Blueprints []BlueprintID `xml:"BlueprintID"`
}

type BlueprintID struct {
	XMLName      xml.Name       `xml:"BlueprintID"`
	BlueprintLib []AbsBluePrint `xml:"iBlueprintLibrary-cAbsoluteBlueprintID"`
}

type AbsBluePrint struct {
	XMLName     xml.Name `xml:"iBlueprintLibrary-cAbsoluteBlueprintID"`
	BlueprintID string   `xml:"BlueprintID"`
}

func main() {
	xmlStruct := RecordSet{}
	if err := serzExe.Run(); err != nil {
		fmt.Println("Error: ", err)
	}

	xmlFile, err := os.Open("./files/" + testFile + ".xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(xmlFile)
	info, err := os.Lstat("./files/" + testFile + ".xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("The size of the xml is: ", info.Size())
	xmlBytes := make([]byte, info.Size())
	numRead, err := xmlFile.Read(xmlBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Number of bytes read: ", numRead)

	err = xml.Unmarshal(xmlBytes, &xmlStruct)
	if err != nil {
		log.Fatal(err)
	}
	assetMap := make(map[string]int)
	entityCount := len(xmlStruct.Records[0].Entities)
	for i := 0; i < entityCount; i++ {
		asset := xmlStruct.Records[0].Entities[i].Blueprints[0].BlueprintLib[0].BlueprintID
		if assetMap[asset] == 0 {
			fmt.Println(asset)
			assetMap[asset]++
		}
	}
}
