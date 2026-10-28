package settings_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"node-herder/mocks"
	"node-herder/models/settings"
	"node-herder/repository"
)

type homeCacheRepo struct {
	mocks.NopSettingsrepo
	config *settings.AppConfig
	fail   bool
	writes int
}

func (r *homeCacheRepo) Load() (*settings.AppConfig, error) {
	encoded, _ := json.Marshal(r.config)
	var config settings.AppConfig
	err := json.Unmarshal(encoded, &config)
	return &config, err
}

func (r *homeCacheRepo) SaveAppConfig(config *settings.AppConfig) error {
	if r.fail {
		return errors.New("private raw storage reason")
	}
	encoded, _ := json.Marshal(config)
	r.writes++
	return json.Unmarshal(encoded, &r.config)
}

func homeCache(t *testing.T) (*settings.AppConfigCache, *homeCacheRepo) {
	t.Helper()
	config := settings.NewAppConfig()
	config.Bridge = settings.DefaultBridgeConfig()
	config.Hub.DashboardGroups = map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen")}
	repo := &homeCacheRepo{config: config}
	cache, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cache, repo
}

func addHomeWaste(layout *settings.HomeLayout) {
	layout.Blocks[0].Cards = append(layout.Blocks[0].Cards, settings.HomeCard{ID: "private-bin", Kind: "waste_service", InstanceID: "private-connection", ResourceID: "calendar", ServiceID: "black", Panel: true})
}

func TestHomeCacheWriteFenceFailureAndPublicBoundary(t *testing.T) {
	cache, repo := homeCache(t)
	initial, err := cache.LoadHomeLayout()
	if err != nil || repo.writes != 0 {
		t.Fatal("GET must not persist migration", err, repo.writes)
	}
	proposal := initial.Clone()
	addHomeWaste(&proposal)
	repo.fail = true
	if _, err := cache.SaveHomeLayout(proposal, initial.ETag); err == nil {
		t.Fatal("failed write accepted")
	}
	unchanged, _ := cache.LoadHomeLayout()
	if unchanged.ETag != initial.ETag {
		t.Fatal("failure published a partial layout")
	}
	repo.fail = false
	saved, err := cache.SaveHomeLayout(proposal, initial.ETag)
	if err != nil || saved.Revision != 1 || repo.config.HomeLayout == nil {
		t.Fatal(saved, err)
	}
	proposal.Blocks[0].Cards[0].Expose = "caller mutation"
	saved.Blocks[0].Cards[0].Expose = "response mutation"
	accepted, _ := cache.LoadHomeLayout()
	if accepted.Blocks[0].Cards[0].Expose != "temperature" {
		t.Fatal("accepted layout aliases input/output")
	}
	if _, err = cache.SaveHomeLayout(initial, initial.ETag); err == nil {
		t.Fatal("stale layout overwrote accepted state")
	}
	if _, err = cache.SaveHomeLayout(initial, ""); err == nil {
		t.Fatal("missing revision was accepted")
	}
	checkPublic := func(config *settings.AppConfig, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(config)
		if bytes.Contains(encoded, []byte("private-connection")) || bytes.Contains(encoded, []byte("homeLayout")) {
			t.Fatal("private layout leaked through public configuration", string(encoded))
		}
	}
	checkPublic(cache.LoadAppConfig())
	checkPublic(cache.SaveLoggerConfig(settings.DefaultLoggingConfig()))
	checkPublic(cache.SaveAssistantConfig(settings.DefaultAssistantConfig()))
	checkPublic(cache.SaveHistoryConfig(settings.DefaultHistoryConfig()))
	checkPublic(cache.SaveMCPConfig(settings.DefaultMCPConfig()))
}

func TestHomeCacheLegacyEditsHaveOnePlacementAuthority(t *testing.T) {
	cache, _ := homeCache(t)
	layout, _ := cache.LoadHomeLayout()
	addHomeWaste(&layout)
	saved, err := cache.SaveHomeLayout(layout, layout.ETag)
	if err != nil {
		t.Fatal(err)
	}
	groupID, cardID := saved.Blocks[0].ID, saved.Blocks[0].Cards[0].ID
	if _, err := cache.RenameDashboardGroup("Kitchen", "Lounge"); err != nil {
		t.Fatal(err)
	}
	after, _ := cache.LoadHomeLayout()
	if after.Blocks[0].ID != groupID || after.Blocks[0].Cards[0].ID != cardID {
		t.Fatal("rename changed placement identity")
	}
	group := homeGroup("Lounge")
	group.DeviceGroup["device"].Exposes = []string{"humidity"}
	if err := cache.SaveDashboardGroup(group); err != nil {
		t.Fatal(err)
	}
	after, _ = cache.LoadHomeLayout()
	if len(after.Blocks[0].Cards) != 2 || after.Blocks[0].Cards[1].Kind != "waste_service" {
		t.Fatal("device edit erased an integration placement", after)
	}
	if err := cache.ImportDashboardGroups(map[string]*settings.DashboardGroup{"Garden": homeGroup("Garden")}); err != nil {
		t.Fatal(err)
	}
	after, _ = cache.LoadHomeLayout()
	if len(after.Blocks) != 2 || after.Blocks[1].Kind != "card" || after.Blocks[1].Cards[0].ID != "private-bin" {
		t.Fatal("legacy import must preserve displaced integration placements", after)
	}
	if err := cache.DeleteDashboardGroup("Garden"); err != nil {
		t.Fatal(err)
	}
	after, _ = cache.LoadHomeLayout()
	if len(after.Blocks) != 1 || after.Blocks[0].Cards[0].ID != "private-bin" {
		t.Fatal("legacy delete changed unrelated integration placement", after)
	}
}

func TestHomeCacheConcurrentSavesConflictAndRestartPersists(t *testing.T) {
	cache, _ := homeCache(t)
	layout, _ := cache.LoadHomeLayout()
	results := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := cache.SaveHomeLayout(layout, layout.ETag)
			results <- err
		}()
	}
	group.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatal("concurrent saves did not fence", winners)
	}
	path := t.TempDir() + "/settings.db"
	repo, err := repository.NewFileSettingsRepoFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	persistent, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	empty, _ := persistent.LoadHomeLayout()
	if _, err = persistent.SaveHomeLayout(layout, empty.ETag); err != nil {
		t.Fatal(err)
	}
	if err = repo.Close(); err != nil {
		t.Fatal(err)
	}
	repo, err = repository.NewFileSettingsRepoFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	restarted, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := restarted.LoadHomeLayout()
	if err != nil || got.Blocks[0].Cards[0].ID != layout.Blocks[0].Cards[0].ID {
		t.Fatal("layout identity was not persisted", got, err)
	}
}
