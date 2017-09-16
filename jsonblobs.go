package main

type ServerExtendedInfo struct {
	Kernel      string `json:"Kernel"`
	OS          string `json:"OS"`
	Name        string `json:"name"`
	LastUpdated string `json:"Last Updated"`
}

type ServerItemResponse struct {
	Name string             `json:"name"`
	Date string             `json:"time"`
	Good bool               `json:"good"`
	Info ServerExtendedInfo `json:"info"`
}

type ServerListResponse struct {
	Servers []ServerItemResponse `json:"servers"`
}
