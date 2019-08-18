package models

// Article ...
type Article struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author Author `json:"author"`
}

// Author ...
type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
