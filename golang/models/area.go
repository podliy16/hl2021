package models

import "fmt"

type Area struct {
	PosX  int `json:"posX"`
	PosY  int `json:"posY"`
	SizeX int `json:"sizeX"`
	SizeY int `json:"sizeY"`
}

func (d Area) String() string {
	return fmt.Sprintf("%d_%d_%d_%d", d.PosX, d.PosY, d.SizeX, d.SizeY)
}

type AreaResponse struct {
	Area   Area `json:"area"`
	Amount int  `json:"amount"`
}
