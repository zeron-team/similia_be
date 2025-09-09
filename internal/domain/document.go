package domain

type Document struct {
	ID                string `json:"id"`
	Folder            string `json:"folder"`
	Filename          string `json:"filename"`
	OriginalFilename  string `json:"originalFilename"`
	Size              int64  `json:"size"`
	Ext               string `json:"ext"`
	UpdatedAt         string `json:"updatedAt"`
	TextContent       string `json:"textContent"`
}
