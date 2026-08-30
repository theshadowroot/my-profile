package services

import (
	"encoding/json"
	"fmt"
	"os"

	"portfolio-os/internal/models"

	"gorm.io/gorm"
)

func MigratePortfolioData(db *gorm.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read portfolio file: %w", err)
	}

	var portfolio models.Portfolio

	if err := json.Unmarshal(data, &portfolio); err != nil {
		return fmt.Errorf("failed to decode portfolio file: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if portfolio.Profile.Name != "" {
			if err := tx.Where("id = ?", portfolio.Profile.ID).
				Assign(portfolio.Profile).
				FirstOrCreate(&portfolio.Profile).Error; err != nil {
				return fmt.Errorf("failed to migrate profile: %w", err)
			}
		}

		for _, statistic := range portfolio.Statistics {
			statistic.ID = 0

			if err := tx.Create(&statistic).Error; err != nil {
				return fmt.Errorf("failed to migrate statistic: %w", err)
			}
		}

		for _, skill := range portfolio.Skills {
			skill.ID = 0

			if err := tx.Create(&skill).Error; err != nil {
				return fmt.Errorf("failed to migrate skill: %w", err)
			}
		}

		for _, service := range portfolio.Services {
			service.ID = 0

			if err := tx.Create(&service).Error; err != nil {
				return fmt.Errorf("failed to migrate service: %w", err)
			}
		}

		for _, education := range portfolio.Education {
			education.ID = 0

			if err := tx.Create(&education).Error; err != nil {
				return fmt.Errorf("failed to migrate education: %w", err)
			}
		}

		for _, certificate := range portfolio.Certificates {
			certificate.ID = 0

			if err := tx.Create(&certificate).Error; err != nil {
				return fmt.Errorf("failed to migrate certificate: %w", err)
			}
		}

		for _, project := range portfolio.Projects {
			project.ID = 0

			if err := tx.Create(&project).Error; err != nil {
				return fmt.Errorf("failed to migrate project: %w", err)
			}
		}

		if portfolio.Contact.Email != "" ||
			portfolio.Contact.Phone != "" ||
			portfolio.Contact.Location != "" {

			portfolio.Contact.ID = 0

			if err := tx.Create(&portfolio.Contact).Error; err != nil {
				return fmt.Errorf("failed to migrate contact: %w", err)
			}
		}

		for _, socialLink := range portfolio.SocialLinks {
			socialLink.ID = 0

			if err := tx.Create(&socialLink).Error; err != nil {
				return fmt.Errorf("failed to migrate social link: %w", err)
			}
		}

		return nil
	})

}
