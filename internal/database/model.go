package database

import "time"

// TableName sets File's table name to be `files`
func (FileModel) TableName() string {
	return "files"
}

type FileModel struct {
	ID          uint64    `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserId      string    `json:"userId"`
	Path        string    `json:"path"`
	Filename    string    `json:"filename"`
	Bucket      string    `json:"-"`
	Storage     string    `json:"storage"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	IsAdminOnly bool      `json:"isAdminOnly"`
	IsPrivate   bool      `json:"isPrivate"`
	Category    *string   `json:"-"`
}
