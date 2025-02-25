package repository

import (
	"testing"
	"time"

	core "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

type testTarget struct {
	name   string
	target model.UserTarget
}

func createTestTargets() []testTarget {
	return []testTarget{
		{
			name: "target with minimal interval",
			target: model.UserTarget{
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
			name: "target with medium interval",
			target: model.UserTarget{
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
			name: "target with large interval",
			target: model.UserTarget{
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

func setupTestTargets(t *testing.T, repo *TargetRepository) map[string]model.UserTarget {
	createdTargets := make(map[string]model.UserTarget)
	for _, tc := range createTestTargets() {
		created, err := repo.Create(tc.target)
		if err != nil {
			t.Fatalf("Failed to create test target %s: %v", tc.name, err)
		}
		createdTargets[tc.name] = created
	}
	return createdTargets
}

func TestTargetRepository_Create(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewTargetRepository(db)

	tests := []struct {
		name    string
		target  model.UserTarget
		wantErr bool
	}{
		{
			name: "valid target",
			target: model.UserTarget{
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
			name: "invalid target - empty URL",
			target: model.UserTarget{
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
			name: "invalid target - invalid user ID",
			target: model.UserTarget{
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
			newTarget, err := repo.Create(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.target.UserID, newTarget.UserID)
			assert.NotZero(t, newTarget.ID)
			assert.Equal(t, tt.target.URL, newTarget.URL)
			if !tt.target.StatusChangedAt.IsZero() {
				assert.WithinDuration(t, tt.target.StatusChangedAt, newTarget.StatusChangedAt, time.Second)
			}
		})
	}
}

func TestTargetRepository_Update(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewTargetRepository(db)

	tests := []struct {
		name        string
		setupTarget model.UserTarget
		updateFunc  func(*model.UserTarget)
		wantErr     bool
	}{
		{
			name: "update status and enabled",
			setupTarget: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			updateFunc: func(s *model.UserTarget) {
				s.Status = "down"
				s.Enabled = true
			},
			wantErr: false,
		},
		{
			name: "update non-existent target",
			setupTarget: model.UserTarget{
				UserID: 1,
				Target: &core.Target{
					ID:       999,
					URL:      "example.org",
					Status:   "up",
					Enabled:  false,
					Interval: 30 * time.Second,
				},
			},
			updateFunc: func(s *model.UserTarget) {
				s.Status = "down"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var userTarget model.UserTarget
			if tt.name == "update non-existent target" {
				userTarget = tt.setupTarget
			} else {
				created, err := repo.Create(tt.setupTarget)
				if err != nil {
					t.Fatalf("Failed to create test target: %v", err)
				}
				userTarget = created
			}

			tt.updateFunc(&userTarget)
			updated, err := repo.Update(userTarget)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			fetched, err := repo.GetByID(userTarget.ID)
			assert.NoError(t, err)
			assert.Equal(t, updated, fetched)
		})
	}
}

func TestTargetRepository_Delete(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewTargetRepository(db)

	tests := []struct {
		name        string
		setupTarget model.UserTarget
		targetID    int
		wantErr     bool
	}{
		{
			name: "delete existing target",
			setupTarget: model.UserTarget{
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
			name:     "delete non-existent target",
			targetID: 999,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetID int
			if tt.setupTarget.Target != nil {
				created, err := repo.Create(tt.setupTarget)
				if err != nil {
					t.Fatalf("Failed to create test target: %v", err)
				}
				targetID = created.ID
			} else {
				targetID = tt.targetID
			}

			err := repo.Delete(targetID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			_, err = repo.GetByID(targetID)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrTargetNotFound)
		})
	}
}

func TestTargetRepository_GetAll(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewTargetRepository(db)

	createdTargets := setupTestTargets(t, repo)

	targets, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, targets, len(createdTargets))

	targetMap := make(map[int]model.UserTarget)
	for _, t := range targets {
		targetMap[t.ID] = t
	}

	for name, created := range createdTargets {
		found, exists := targetMap[created.ID]
		assert.True(t, exists, "Target '%s' should exist", name)
		if exists {
			assert.Equal(t, created.URL, found.URL, "URL mismatch for %s", name)
			assert.Equal(t, created.Status, found.Status, "Status mismatch for %s", name)
			assert.Equal(t, created.Enabled, found.Enabled, "Enabled mismatch for %s", name)
			assert.Equal(t, created.Interval, found.Interval, "Interval mismatch for %s", name)
		}
	}
}

func TestTargetRepository_GetAllByUserID(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()
	repo := NewTargetRepository(db)

	setupTestTargets(t, repo)

	tests := []struct {
		name          string
		userID        int
		expectedURLs  []string
		expectedCount int
	}{
		{
			name:          "user with multiple targets",
			userID:        1,
			expectedURLs:  []string{"example1.org", "example2.org"},
			expectedCount: 2,
		},
		{
			name:          "user with single target",
			userID:        2,
			expectedURLs:  []string{"example3.org"},
			expectedCount: 1,
		},
		{
			name:          "user with no targets",
			userID:        999,
			expectedURLs:  []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets, err := repo.GetAllByUserID(tt.userID)
			assert.NoError(t, err)
			assert.Len(t, targets, tt.expectedCount)

			if tt.expectedCount > 0 {
				urls := make([]string, len(targets))
				for i, target := range targets {
					urls[i] = target.URL
				}
				for _, expectedURL := range tt.expectedURLs {
					assert.Contains(t, urls, expectedURL)
				}
			}
		})
	}
}
