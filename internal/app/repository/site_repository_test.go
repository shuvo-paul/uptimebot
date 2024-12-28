package repository

import (
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/testutil"
	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
	"github.com/stretchr/testify/assert"
)

func TestSiteRepository(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	siteRepo := NewSiteRepository(db)

	createTestSite := func() *monitor.Site {
		return &monitor.Site{
			URL:             "example.org",
			Status:          "up",
			Enabled:         false,
			Interval:        30 * time.Second,
			StatusChangedAt: time.Now(),
		}
	}

	t.Run("create", func(t *testing.T) {
		tests := []struct {
			name    string
			site    *monitor.Site
			wantErr bool
		}{
			{
				name:    "valid site",
				site:    createTestSite(),
				wantErr: false,
			},
			{
				name: "invalid site - empty URL",
				site: &monitor.Site{
					URL:      "",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				newSite, err := siteRepo.Create(tt.site)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotZero(t, newSite.ID)
				assert.Equal(t, tt.site.URL, newSite.URL)
				assert.WithinDuration(t, tt.site.StatusChangedAt, newSite.StatusChangedAt, time.Second)
			})
		}
	})

	t.Run("update", func(t *testing.T) {
		site := createTestSite()
		created, err := siteRepo.Create(site)
		assert.NoError(t, err)

		created.Status = "down"
		created.Enabled = true
		updated, err := siteRepo.Update(created)
		assert.NoError(t, err)
		assert.Equal(t, "down", updated.Status)
		assert.True(t, updated.Enabled)

		fetched, err := siteRepo.GetByID(created.ID)
		assert.NoError(t, err)
		assert.Equal(t, updated, fetched)
	})

	t.Run("delete", func(t *testing.T) {
		site := createTestSite()
		created, err := siteRepo.Create(site)
		assert.NoError(t, err)

		err = siteRepo.Delete(created.ID)
		assert.NoError(t, err)

		_, err = siteRepo.GetByID(created.ID)
		assert.Error(t, err)
	})

	t.Run("get all", func(t *testing.T) {
		testSites := []struct {
			name string
			site *monitor.Site
		}{
			{
				name: "site with minimal interval",
				site: &monitor.Site{
					URL:             "example1.org",
					Status:          "up",
					Enabled:         true,
					Interval:        30 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
			{
				name: "site with medium interval",
				site: &monitor.Site{
					URL:             "example2.org",
					Status:          "down",
					Enabled:         false,
					Interval:        60 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
			{
				name: "site with large interval",
				site: &monitor.Site{
					URL:             "example3.org",
					Status:          "up",
					Enabled:         true,
					Interval:        90 * time.Second,
					StatusChangedAt: time.Now(),
				},
			},
		}

		createdSites := make(map[string]*monitor.Site)
		for _, tc := range testSites {
			created, err := siteRepo.Create(tc.site)
			assert.NoError(t, err)
			createdSites[tc.name] = created
		}

		sites, err := siteRepo.GetAll()
		assert.NoError(t, err)

		assert.GreaterOrEqual(t, len(sites), len(testSites))

		siteMap := make(map[int]*monitor.Site)
		for _, s := range sites {
			siteMap[s.ID] = s
		}

		for _, tc := range testSites {
			created := createdSites[tc.name]
			found, exists := siteMap[created.ID]
			assert.True(t, exists, "Site '%s' should exist", tc.name)
			if exists {
				assert.Equal(t, created.URL, found.URL, "URL mismatch for %s", tc.name)
				assert.Equal(t, created.Status, found.Status, "Status mismatch for %s", tc.name)
				assert.Equal(t, created.Enabled, found.Enabled, "Enabled mismatch for %s", tc.name)
				assert.Equal(t, created.Interval, found.Interval, "Interval mismatch for %s", tc.name)
			}
		}
	})
}
