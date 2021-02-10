package types

import "fmt"

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

type DotCounter struct {
	count int
}

func (d *DotCounter) PrintDot() {
	if d.count%50 == 0 {
		fmt.Printf(".")
	}
	d.count++
}

func NewDotCounter() *DotCounter {
	return &DotCounter{
		count: 0,
	}
}
