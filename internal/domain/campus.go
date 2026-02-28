package domain

import "context"

type University struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	City   string `json:"city"`
}

type CampusLocation struct {
	ID           int     `json:"id"`
	UniversityID int     `json:"university_id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
}

type CampusRepository interface {
	GetUniversities(ctx context.Context) ([]*University, error)
	GetLocations(ctx context.Context, universityID int) ([]*CampusLocation, error)
}
