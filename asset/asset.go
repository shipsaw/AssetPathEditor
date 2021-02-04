package asset

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Asset struct {
	Product  string
	Provider string
	Filepath string
}

var EmptyAsset = Asset{
	Product:  "",
	Provider: "",
	Filepath: "",
}

type AssetAssetMap map[Asset]Asset
type AssetBoolMap map[Asset]bool

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
	fmt.Printf("\n\nRoute Dependancies:\n")
	fmt.Printf("%-19v%-19v\n", "Provider", "Product")
	fmt.Println("--------------------------------------")
	for _, ProvProd := range assetList {
		fmt.Printf("%-19v%-19v\n", ProvProd[0], ProvProd[1])
	}
	fmt.Printf("\n\n")
	return uniqueAssets
}

func Check(misAssets AssetAssetMap, allAssets AssetBoolMap) {
	fmt.Printf("There are initially %d required assets\n", len(misAssets))
	rightPlace := 0
	differentPlace := 0
	notFound := 0
OUTER:
	for misAsset, value := range misAssets {
		for locAsset, _ := range allAssets {
			_, ok := allAssets[misAsset]
			if ok == true {
				rightPlace++
				delete(misAssets, misAsset)
				continue OUTER
			} else if misAsset.Filepath == locAsset.Filepath {
				tempAsset := Asset{
					Product:  locAsset.Product,
					Provider: locAsset.Provider,
					Filepath: locAsset.Filepath,
				}
				misAssets[misAsset] = tempAsset
				differentPlace++
				continue OUTER
			}
		}
		if value == EmptyAsset {
			fmt.Println(misAsset)
			notFound++
		}
	}
	fmt.Printf("%v assets cannot be found\n", notFound)
	fmt.Printf("%v assets have been found in the correct folder\n", rightPlace)
	fmt.Printf("%v assets have been found, but not in the requested location\n", differentPlace)
}

func Index(misAssets AssetAssetMap) (AssetBoolMap, error) {
	allAssets := make(AssetBoolMap)
	err := filepath.Walk(`.\Assets`, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".bin" {
			// Seperate to find providers, products, and paths
			pathSlice := strings.SplitN(path, `\`, 4)
			if len(pathSlice) == 4 {
				asset := Asset{
					Product:  pathSlice[2],
					Provider: pathSlice[1],
					Filepath: pathSlice[3],
				}
				allAssets[asset] = false
			}
		} else if filepath.Ext(path) == ".ap" {
			GetZipAssets(path, misAssets, allAssets)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allAssets, nil
}

func GetZipAssets(filename string, misAssets AssetAssetMap, allAssets AssetBoolMap) error {
	filenameSlice := strings.SplitN(filename, `\`, 4)
	var buf bytes.Buffer
	cmd := exec.Command("7z.exe", "l", filename, "-ba")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return err
	}
	zipString := buf.String()
	for misAsset, _ := range misAssets {
		if strings.Contains(zipString, misAsset.Filepath) {
			asset := Asset{
				Product:  filenameSlice[2],
				Provider: filenameSlice[1],
				Filepath: misAsset.Filepath,
			}
			allAssets[asset] = true
		}
	}
	return nil
}
