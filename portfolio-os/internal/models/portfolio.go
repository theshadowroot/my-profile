package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}

	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return string(data), nil

}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported StringArray value type: %T", value)
	}

	if len(data) == 0 {
		*s = []string{}
		return nil
	}

	return json.Unmarshal(data, s)

}

type Portfolio struct {
	Profile      Profile
	Statistics   []Statistic
	Skills       []Skill
	Services     []ServiceOffering
	Education    []EducationEntry
	Certificates []Certificate
	Projects     []Project
	Contact      ContactInfo
	SocialLinks  []SocialLink
}

type Profile struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	Title        string
	Bio          string `gorm:"type:text"`
	Location     string
	ProfileImage string
	Availability string
	CreatedAt    int64
	UpdatedAt    int64
}

type Statistic struct {
	ID        uint   `gorm:"primaryKey"`
	Label     string `gorm:"not null"`
	Value     string `gorm:"not null"`
	CreatedAt int64
	UpdatedAt int64
}

type Skill struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Category  string
	CreatedAt int64
	UpdatedAt int64
}

type ServiceOffering struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Icon        string
	CreatedAt   int64
	UpdatedAt   int64
}

type EducationEntry struct {
	ID          uint   `gorm:"primaryKey"`
	Institution string `gorm:"not null"`
	Degree      string
	Field       string
	StartYear   int
	EndYear     int
	Description string `gorm:"type:text"`
	CreatedAt   int64
	UpdatedAt   int64
}

type Certificate struct {
	ID              uint   `gorm:"primaryKey"`
	Title           string `gorm:"not null"`
	Issuer          string
	IssueDate       string
	CredentialID    string
	Description     string `gorm:"type:text"`
	Image           string
	VerificationURL string
	CreatedAt       int64
	UpdatedAt       int64
}

type Project struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	Description  string `gorm:"type:text"`
	Image        string
	URL          string
	GitHub       string
	Technologies StringArray `gorm:"type:jsonb"`
	CreatedAt    int64
	UpdatedAt    int64
}

type ContactInfo struct {
	ID        uint `gorm:"primaryKey"`
	Email     string
	Phone     string
	Location  string
	CreatedAt int64
	UpdatedAt int64
}

type SocialLink struct {
	ID        uint   `gorm:"primaryKey"`
	Platform  string `gorm:"not null"`
	URL       string
	Username  string
	CreatedAt int64
	UpdatedAt int64
}
