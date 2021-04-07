package models

type Dig struct {
	LicenseID int `json:"licenseID"`
	PosX      int `json:"posX"`
	PosY      int `json:"posY"`
	Depth     int `json:"depth"`
}
