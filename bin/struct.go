package bin

import "encoding/xml"

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
