package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"portfolio-os/internal/models"
)

const analyticsStoragePath = "storage/analytics.json"

type AnalyticsService struct {
	mu     sync.RWMutex
	visits map[string]*models.Visit
}
type AnalyticsStats struct {
	TotalVisits         int             `json:"total_visits"`
	UniqueVisitors      int             `json:"unique_visitors"`
	AverageDuration     time.Duration   `json:"-"`
	AverageDurationText string          `json:"average_duration"`
	DeviceBreakdown     map[string]int  `json:"device_breakdown"`
	PageViews           map[string]int  `json:"page_views"`
	RecentVisits        []*models.Visit `json:"recent_visits"`
}

func NewAnalyticsService() (*AnalyticsService, error) {
	s := &AnalyticsService{
		visits: make(map[string]*models.Visit),
	}

	if err := s.loadLocked(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *AnalyticsService) loadLocked() error {
	if _, err := os.Stat(analyticsStoragePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(analyticsStoragePath)
	if err != nil {
		return fmt.Errorf("open analytics storage: %w", err)
	}
	defer file.Close()

	var visits []*models.Visit

	if err := json.NewDecoder(file).Decode(&visits); err != nil {
		return fmt.Errorf("decode analytics storage: %w", err)
	}

	for _, visit := range visits {
		if visit != nil {
			s.visits[visit.ID] = visit
		}
	}

	return nil
}

func (s *AnalyticsService) StartVisit(visit *models.Visit) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.visits[visit.ID] = visit

	if err := s.saveLocked(); err != nil {
		log.Printf("analytics: save after start visit: %v", err)
	}
}

func (s *AnalyticsService) UpdateDuration(
	visitID string,
	duration time.Duration,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	visit, exists := s.visits[visitID]

	if !exists {
		return false
	}

	visit.Duration = duration

	if err := s.saveLocked(); err != nil {
		log.Printf("analytics: save after duration update: %v", err)
	}

	return true
}

func (s *AnalyticsService) GetVisits() []*models.Visit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visits := make([]*models.Visit, 0, len(s.visits))

	for _, visit := range s.visits {
		visits = append(visits, visit)
	}

	return visits
}

func (s *AnalyticsService) GetStats() *AnalyticsStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &AnalyticsStats{
		TotalVisits:     len(s.visits),
		DeviceBreakdown: make(map[string]int),
		PageViews:       make(map[string]int),
	}

	uniqueVisitors := make(map[string]struct{})

	var totalDuration time.Duration

	visits := make([]*models.Visit, 0, len(s.visits))

	for _, visit := range s.visits {
		visits = append(visits, visit)

		// Unique visitor based on anonymized IP.
		uniqueVisitors[visit.IPAddress] = struct{}{}

		// Device statistics.
		stats.DeviceBreakdown[visit.DeviceCategory]++

		// Page statistics.
		stats.PageViews[visit.Page]++

		// Duration.
		totalDuration += visit.Duration
	}

	stats.UniqueVisitors = len(uniqueVisitors)

	if stats.TotalVisits > 0 {
		stats.AverageDuration = totalDuration / time.Duration(stats.TotalVisits)
		stats.AverageDurationText = FormatDuration(stats.AverageDuration)
	}

	// Most recent visits first.
	sort.Slice(visits, func(i, j int) bool {
		return visits[i].Timestamp.After(visits[j].Timestamp)
	})

	// Return only the latest 10 visits.
	if len(visits) > 10 {
		visits = visits[:10]
	}

	stats.RecentVisits = visits

	return stats
}

func (s *AnalyticsService) saveLocked() error {
	if err := os.MkdirAll("storage", 0755); err != nil {
		return fmt.Errorf("create analytics storage directory: %w", err)
	}

	visits := make([]*models.Visit, 0, len(s.visits))

	for _, visit := range s.visits {
		visits = append(visits, visit)
	}

	sort.Slice(visits, func(i, j int) bool {
		return visits[i].Timestamp.Before(visits[j].Timestamp)
	})

	file, err := os.Create(analyticsStoragePath)
	if err != nil {
		return fmt.Errorf("create analytics storage: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(visits); err != nil {
		return fmt.Errorf("encode analytics storage: %w", err)
	}

	return nil
}

func FormatDuration(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())

	if totalSeconds < 60 {
		return fmt.Sprintf("%d sec", totalSeconds)
	}

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	if minutes < 60 {
		if seconds == 0 {
			return fmt.Sprintf("%d min", minutes)
		}

		return fmt.Sprintf("%d min %d sec", minutes, seconds)
	}

	hours := minutes / 60
	minutes %= 60

	if minutes == 0 {
		return fmt.Sprintf("%d hr", hours)
	}

	return fmt.Sprintf("%d hr %d min", hours, minutes)
}
