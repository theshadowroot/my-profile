package services

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"portfolio-os/internal/models"
)

type PortfolioService struct {
	mu        sync.RWMutex
	filePath  string
	portfolio models.Portfolio
}

func NewPortfolioService(filePath string) (*PortfolioService, error) {
	s := &PortfolioService{
		filePath: filePath,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *PortfolioService) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.portfolio)
}

func (s *PortfolioService) save() error {
	data, err := json.MarshalIndent(s.portfolio, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *PortfolioService) GetPortfolio() models.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portfolio
}

// Profile
func (s *PortfolioService) UpdateProfile(p models.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Profile = p
	return s.save()
}

// Certificates
func (s *PortfolioService) AddCertificate(c models.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Certificates = append(s.portfolio.Certificates, c)
	return s.save()
}

func (s *PortfolioService) UpdateCertificate(idx int, c models.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Certificates) {
		return errors.New("certificate index out of range")
	}
	s.portfolio.Certificates[idx] = c
	return s.save()
}

func (s *PortfolioService) DeleteCertificate(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Certificates) {
		return errors.New("certificate index out of range")
	}
	s.portfolio.Certificates = append(s.portfolio.Certificates[:idx], s.portfolio.Certificates[idx+1:]...)
	return s.save()
}

// Skills
func (s *PortfolioService) AddSkill(skill models.Skill) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Skills = append(s.portfolio.Skills, skill)
	return s.save()
}

func (s *PortfolioService) UpdateSkill(idx int, skill models.Skill) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Skills) {
		return errors.New("skill index out of range")
	}
	s.portfolio.Skills[idx] = skill
	return s.save()
}

func (s *PortfolioService) DeleteSkill(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Skills) {
		return errors.New("skill index out of range")
	}
	s.portfolio.Skills = append(s.portfolio.Skills[:idx], s.portfolio.Skills[idx+1:]...)
	return s.save()
}

// Statistics
func (s *PortfolioService) AddStatistic(stat models.Statistic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Statistics = append(s.portfolio.Statistics, stat)
	return s.save()
}

func (s *PortfolioService) UpdateStatistic(idx int, stat models.Statistic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Statistics) {
		return errors.New("statistic index out of range")
	}
	s.portfolio.Statistics[idx] = stat
	return s.save()
}

func (s *PortfolioService) DeleteStatistic(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Statistics) {
		return errors.New("statistic index out of range")
	}
	s.portfolio.Statistics = append(s.portfolio.Statistics[:idx], s.portfolio.Statistics[idx+1:]...)
	return s.save()
}

// Services
func (s *PortfolioService) AddService(srv models.ServiceOffering) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Services = append(s.portfolio.Services, srv)
	return s.save()
}

func (s *PortfolioService) UpdateService(idx int, srv models.ServiceOffering) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Services) {
		return errors.New("service index out of range")
	}
	s.portfolio.Services[idx] = srv
	return s.save()
}

func (s *PortfolioService) DeleteService(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Services) {
		return errors.New("service index out of range")
	}
	s.portfolio.Services = append(s.portfolio.Services[:idx], s.portfolio.Services[idx+1:]...)
	return s.save()
}

// Education
func (s *PortfolioService) AddEducation(edu models.EducationEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Education = append(s.portfolio.Education, edu)
	return s.save()
}

func (s *PortfolioService) UpdateEducation(idx int, edu models.EducationEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Education) {
		return errors.New("education index out of range")
	}
	s.portfolio.Education[idx] = edu
	return s.save()
}

func (s *PortfolioService) DeleteEducation(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Education) {
		return errors.New("education index out of range")
	}
	s.portfolio.Education = append(s.portfolio.Education[:idx], s.portfolio.Education[idx+1:]...)
	return s.save()
}

// Projects
func (s *PortfolioService) AddProject(proj models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Projects = append(s.portfolio.Projects, proj)
	return s.save()
}

func (s *PortfolioService) UpdateProject(idx int, proj models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Projects) {
		return errors.New("project index out of range")
	}
	s.portfolio.Projects[idx] = proj
	return s.save()
}

func (s *PortfolioService) DeleteProject(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.Projects) {
		return errors.New("project index out of range")
	}
	s.portfolio.Projects = append(s.portfolio.Projects[:idx], s.portfolio.Projects[idx+1:]...)
	return s.save()
}

// Contact & Social
func (s *PortfolioService) UpdateContact(contact models.ContactInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.Contact = contact
	return s.save()
}

func (s *PortfolioService) AddSocialLink(link models.SocialLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portfolio.SocialLinks = append(s.portfolio.SocialLinks, link)
	return s.save()
}

func (s *PortfolioService) UpdateSocialLink(idx int, link models.SocialLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.SocialLinks) {
		return errors.New("social link index out of range")
	}
	s.portfolio.SocialLinks[idx] = link
	return s.save()
}

func (s *PortfolioService) DeleteSocialLink(idx int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.portfolio.SocialLinks) {
		return errors.New("social link index out of range")
	}
	s.portfolio.SocialLinks = append(s.portfolio.SocialLinks[:idx], s.portfolio.SocialLinks[idx+1:]...)
	return s.save()
}