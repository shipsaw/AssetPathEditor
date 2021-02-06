package main

import (
	"log"
	"trainTest/asset"
)

const (
	routeFolder  string = `.\Content\Routes\89f87a1c-fbd4-4f05-ba8b-16069484fa41\`
	backupFolder string = `AssetBackup\`
	replaceRoute string = `.\Content\Routes\3a99321a-0bb2-47be-bcad-b20cfe48a945\`
)

func main() {
	providers, err := asset.ListProviders(replaceRoute)
	if err != nil {
		log.Fatal(err)
	}
	/*
		providers := asset.ProviderMap{
			"Foliage01":         "RSDL",
			"IslandLine":        "RSDL",
			"GEML":              "RSC",
			"APStation":         "RSC",
			"RailSimulatorUS":   "Kuju",
			"RailSimulatorCore": "Kuju",
			"RailSimulator":     "Kuju",
			"WherryLines":       "AP",
		}
	*/
	err = asset.UpdateRoute(routeFolder, providers)
	if err != nil {
		log.Fatal(err)
	}
}
