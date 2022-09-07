package neo

type Link struct {
	Type string `json:"type"`
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Meta Meta   `json:"meta"`
}

type Meta struct {
	ID string `json:"id"`
}
