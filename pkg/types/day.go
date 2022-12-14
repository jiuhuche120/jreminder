package types

type Day struct {
	Code    int     `json:"code,omitempty"`
	Type    Type    `json:"types"`
	Holiday Holiday `json:"holiday"`
}

type Type struct {
	Type int    `json:"types,omitempty"`
	Name string `json:"name,omitempty"`
	Week int    `json:"week,omitempty"`
}

type Holiday struct {
	Holiday bool   `json:"holiday,omitempty"`
	Name    string `json:"name,omitempty"`
	Wage    int    `json:"wage,omitempty"`
	After   bool   `json:"after,omitempty"`
	Target  string `json:"target,omitempty"`
}
