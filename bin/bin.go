package bin

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"trainTest/asset"
)

var (
	ErrCopyMismatch error = errors.New("error: file copy size mismatch")
)

func Setup(binFolder, backupFolder string) error {
	if err := os.Mkdir("tempFiles", 0755); err != nil {
		return err
	}
	if err := os.Mkdir(backupFolder, 0755); err != nil {
		return err
	}
	if err := backupScenery(binFolder, backupFolder); err != nil {
		return err
	}
	return nil
}

func Teardown(backupFolder string, removeBackups bool) {
	if removeBackups == true {
		os.Remove(backupFolder)
	}
	os.Remove("tempFiles")
}

func backupScenery(srcFolder, dstFolder string) error {
	err := filepath.Walk(srcFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && filepath.Ext(path) == ".bin" {
			origFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer origFile.Close()

			newFile, err := os.Create(dstFolder + info.Name())
			if err != nil {
				return err
			}
			writ, err := io.Copy(newFile, origFile)
			if err != nil {
				return err
			}
			if writ != info.Size() {
				return ErrCopyMismatch
			}
			return newFile.Close()
		}
		return nil
	})
	return err
}

func Revert(binFolder, backupFolder string) error {
	return backupScenery(backupFolder, binFolder)
}

func SerzConvert(binFolder, ext string) error {
	fmt.Printf("Converting files")
	err := filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && filepath.Ext(path) == ext {
			cmd := exec.Command("./serz.exe", path)
			if err := cmd.Run(); err != nil {
				return err
			}
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		fmt.Printf(".")
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}

func ListReqAssets(binFolder string) (asset.MisAssetMap, error) {
	err := SerzConvert(binFolder, ".bin")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Processing xml files")
	misAssetMap := make(asset.MisAssetMap)
	err = filepath.Walk(binFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && filepath.Ext(path) == ".xml" {
			err = getFileAssets(path, misAssetMap)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n")
	return misAssetMap, nil
}

func getFileAssets(path string, misAssets asset.MisAssetMap) error {
	fmt.Printf(".")
	xmlStruct := RecordSet{}
	// Open xml file
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
	err = xml.Unmarshal(xmlBytes, &xmlStruct)
	if err != nil {
		return err
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
	return nil
}

func MoveXmlFiles(oldLoc string, newLoc string) error {
	err := filepath.Walk(oldLoc, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() != true && filepath.Ext(path) == ".xml" {
			newPath := newLoc + filepath.Base(path)
			err = os.Rename(path, newPath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ReplaceXmlText(xmlFolder string, misAssets asset.MisAssetMap) error {
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
			defer xmlFile.Close()
			fileBytes := make([]byte, info.Size())
			_, err = xmlFile.Read(fileBytes)
			if err != nil {
				return err
			}

			for oldAsset, newAsset := range misAssets {
				if newAsset == asset.EmptyAsset {
					continue
				}
				fixPath := strings.ReplaceAll(oldAsset.Filepath, `\`, `\\`)
				fixPath = strings.ReplaceAll(fixPath, ".bin", ".xml")
				fixNewPath := strings.ReplaceAll(newAsset.Filepath, ".bin", ".xml")
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
				retRegNew = bytes.Replace(retRegNew, matches[3], []byte(fixNewPath), 1)
				fileBytes = bytes.Replace(fileBytes, retReg, retRegNew, -1)

			}
			fmt.Printf(".")
			err = xmlFile.Truncate(0)
			if err != nil {
				return err
			}
			_, err = xmlFile.WriteAt(fileBytes, 0)
			if err != nil {
				return err
			}

		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}
