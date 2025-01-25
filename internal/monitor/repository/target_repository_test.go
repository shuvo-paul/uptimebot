package repository

import (
	"testing"
	"time"

	core "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

type testSite struct {
	name string
	site model.UserTarget
}

func createTestSites() []testSite {
	return []testSite{
		{
			name: "site with minimal interval",
			site: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:             "example1.org",
					Status:          "up",
					Enabled:         true,
					Interval:        30 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
		},
		{
			name: "site with medium interval",
			site: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:             "example2.org",
					Status:          "down",
					Enabled:         false,
					Interval:        60 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
		},
		{
			name: "site with large interval",
			site: model.UserTarget{
				UserID: 2,
				Target: &core.Target{
					URL:             "example3.org",
					Status:          "up",
					Enabled:         true,
					Interval:        90 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
		},
	}
}

func setupTestSites(t *testing.T, repo *SiteRepository) map[string]model.UserTarget {
	createdSites := make(map[string]model.UserTarget)
	for _, tc := range createTestSites() {
		created, err := repo.Create(tc.site)
		if err != nil {
			t.Fatalf("Failed to create test site %s: %v", tc.name, err)
		}
		createdSites[tc.name] = created
	}
	return createdSites
}

func TestSiteRepository_Create(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewSiteRepository(db)

	tests := []struct {
		name    string
		site    model.UserTarget
		wantErr bool
	}{
		{
			name: "valid site",
			site: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:             "example.org",
					Status:          "up",
					Enabled:         false,
					Interval:        30 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid site - empty URL",
			site: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:      "",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid site - invalid user ID",
			site: model.UserTarget{
				UserID: 0,
				Target: &core.Target{
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newSite, err := repo.Create(tt.site)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.site.UserID, newSite.UserID)
			assert.NotZero(t, newSite.ID)
			assert.Equal(t, tt.site.URL, newSite.URL)
			if !tt.site.StatusChangedAt.IsZero() {
				assert.WithinDuration(t, tt.site.StatusChangedAt, newSite.StatusChangedAt, time.Second)
			}
		})
	}
}

func TestSiteRepository_Update(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewSiteRepository(db)

	tests := []struct {
		name       string
		setupSite  model.UserTarget
		updateFunc func(*core.Target)
		wantErr    bool
	}{
		{
			name: "update status and enabled",
			setupSite: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			updateFunc: func(s *core.Target) {
				s.Status = "down"
				s.Enabled = true
			},
			wantErr: false,
		},
		{
			name: "update non-existent site",
			setupSite: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					ID:       999,
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			updateFunc: func(s *core.Target) {
				s.Status = "down"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var site *core.Target
			if tt.name == "update non-existent site" {
				site = tt.setupSite.Target
			} else {
				created, err := repo.Create(tt.setupSite)
				if err != nil {
					t.Fatalf("Failed to create test site: %v", err)
				}
				site = created.Target
			}

			tt.updateFunc(site)
			updated, err := repo.Update(site)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			fetched, err := repo.GetByID(site.ID)
			assert.NoError(t, err)
			assert.Equal(t, updated, fetched)
		})
	}
}

func TestSiteRepository_Delete(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewSiteRepository(db)

	tests := []struct {
		name      string
		setupSite model.UserTarget
		siteID    int
		wantErr   bool
	}{
		{
			name: "delete existing site",
			setupSite: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent site",
			siteID:  999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var siteID int
			if tt.setupSite.Target != nil {
				created, err := repo.Create(tt.setupSite)
				if err != nil {
					t.Fatalf("Failed to create test site: %v", err)
				}
				siteID = created.ID
			} else {
				siteID = tt.siteID
			}

			err := repo.Delete(siteID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			_, err = repo.GetByID(siteID)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrTargetNotFound)
		})
	}
}

func TestSiteRepository_GetAll(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewSiteRepository(db)

	createdSites := setupTestSites(t, repo)

	sites, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, sites, len(createdSites))

	siteMap := make(map[int]*core.Target)
	for _, s := range sites {
		siteMap[s.ID] = s
	}

	for name, created := range createdSites {
		found, exists := siteMap[created.ID]
		assert.True(t, exists, "Site '%s' should exist", name)
		if exists {
			assert.Equal(t, created.URL, found.URL, "URL mismatch for %s", name)
			assert.Equal(t, created.Status, found.Status, "Status mismatch for %s", name)
			assert.Equal(t, created.Enabled, found.Enabled, "Enabled mismatch for %s", name)
			assert.Equal(t, created.Interval, found.Interval, "Interval mismatch for %s", name)
		}
	}
}

func TestSiteRepository_GetAllByUserID(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewSiteRepository(db)

	setupTestSites(t, repo)

	tests := []struct {
		name          string
		userID        int
		expectedURLs  []string
		expectedCount int
	}{
		{
			name:          "user with multiple sites",
			userID:        1,
			expectedURLs:  []string{"example1.org", "example2.org"},
			expectedCount: 2,
		},
		{
			name:          "user with single site",
			userID:        2,
			expectedURLs:  []string{"example3.org"},
			expectedCount: 1,
		},
		{
			name:          "user with no sites",
			userID:        999,
			expectedURLs:  []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sites, err := repo.GetAllByUserID(tt.userID)
			assert.NoError(t, err)
			assert.Len(t, sites, tt.expectedCount)

			if tt.expectedCount > 0 {
				urls := make([]string, len(sites))
				for i, site := range sites {
					urls[i] = site.URL
				}
				for _, expectedURL := range tt.expectedURLs {
					assert.Contains(t, urls, expectedURL)
				}
			}
		})
	}
}
