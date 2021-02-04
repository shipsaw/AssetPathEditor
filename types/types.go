package types

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
