package settings

import "slices"

func (s *AppConfigCache) commitLegacyGroupsLocked(groups map[string]*DashboardGroup) error {
	if groups == nil || len(groups) > 50 {
		return invalidHome()
	}
	for name, group := range groups {
		if group == nil || group.Name != name {
			return invalidHome()
		}
	}
	candidate := *s.appConfig
	candidate.Hub = cloneHubConfig(s.appConfig.Hub)
	candidate.Hub.DashboardGroups = cloneHubConfig(&HubConfig{DashboardGroups: groups}).DashboardGroups
	if err := s.store.SaveAppConfig(&candidate); err != nil {
		return err
	}
	s.appConfig.Hub = candidate.Hub
	return nil
}

// LoadHomeLayout reads an owned private layout. Legacy migration is deterministic and read-only
// until the first accepted save; device operation never depends on integration storage readiness.
func (s *AppConfigCache) LoadHomeLayout() (HomeLayout, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.homeLayoutLocked()
}

func (s *AppConfigCache) homeLayoutLocked() (HomeLayout, error) {
	var layout HomeLayout
	if s.appConfig.HomeLayout != nil {
		layout = s.appConfig.HomeLayout.Clone()
	} else {
		layout = HomeFromGroups(s.appConfig.Hub.DashboardGroups)
	}
	if layout.Version != 1 {
		return HomeLayout{}, homeSchemaError(layout.Version)
	}
	if err := layout.Validate(); err != nil {
		return HomeLayout{}, err
	}
	layout.ETag = HomeETag(layout)
	return layout, nil
}

// SaveHomeLayout atomically persists the private placement and device-only projection under the
// existing cache lock/Bolt record transaction. Conflicts/failures do not publish proposed state.
func (s *AppConfigCache) SaveHomeLayout(layout HomeLayout, expected string) (HomeLayout, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current, err := s.homeLayoutLocked()
	if err != nil {
		return HomeLayout{}, err
	}
	if expected == "" {
		return HomeLayout{}, homeFailure("precondition_required", "Reload Home before saving its layout.")
	}
	if expected != current.ETag {
		return HomeLayout{}, homeFailure("revision_conflict", "Home changed in another session. Reload the layout and try again.")
	}
	return s.commitHomeLocked(layout, current.Revision)
}

func (s *AppConfigCache) commitHomeLocked(layout HomeLayout, revision uint64) (HomeLayout, error) {
	if err := layout.Validate(); err != nil {
		return HomeLayout{}, err
	}
	layout = layout.Clone()
	layout.Revision = revision + 1
	layout.ETag = HomeETag(layout)
	candidate := *s.appConfig
	candidate.Hub = cloneHubConfig(s.appConfig.Hub)
	candidate.Hub.DashboardGroups = layout.DeviceGroups()
	candidate.HomeLayout = &layout
	if err := s.store.SaveAppConfig(&candidate); err != nil {
		return HomeLayout{}, homeFailure("storage_error", "Home could not be saved. The previous layout is still active.")
	}
	// Keep the AppConfig pointer shared with DeviceConfigCache, publishing only owned fields.
	s.appConfig.Hub = candidate.Hub
	s.appConfig.HomeLayout = &layout
	return layout.Clone(), nil
}

// publicSnapshotLocked is the only AppConfig response boundary. The persisted private layout
// must be omitted even from settings-save replies, not merely the initial hub snapshot.
func (s *AppConfigCache) publicSnapshotLocked() *AppConfig {
	config := *s.appConfig
	config.HomeLayout = nil
	config.Hub = cloneHubConfig(config.Hub)
	config.Bridge = cloneBridgeConfig(config.Bridge)
	return &config
}

func replaceLegacyCards(block HomeBlock, devices map[string]*DeviceGroup) HomeBlock {
	wanted := legacyDeviceCards(block.ID, devices)
	bySource := make(map[string]HomeCard, len(wanted))
	for _, card := range wanted {
		bySource[card.SourceKey()] = card
	}
	block.Cards = slices.Clone(block.Cards)
	next := make([]HomeCard, 0, len(block.Cards)+len(wanted))
	for _, card := range block.Cards {
		if card.Kind != "device" {
			next = append(next, card)
			continue
		}
		if _, exists := bySource[card.SourceKey()]; exists {
			next = append(next, card)
			delete(bySource, card.SourceKey())
		}
	}
	for _, card := range wanted {
		if _, exists := bySource[card.SourceKey()]; exists {
			next = append(next, card)
			delete(bySource, card.SourceKey())
		}
	}
	block.Cards = next
	return block
}

func (s *AppConfigCache) saveLegacyHomeGroupLocked(group *DashboardGroup) error {
	layout, err := s.homeLayoutLocked()
	if err != nil {
		return err
	}
	index := slices.IndexFunc(layout.Blocks, func(block HomeBlock) bool { return block.Kind == "group" && block.Name == group.Name })
	if index < 0 {
		block := HomeBlock{ID: legacyHomeID("group", group.Name), Kind: "group", Name: group.Name, Cards: []HomeCard{}}
		layout.Blocks = append(layout.Blocks, replaceLegacyCards(block, group.DeviceGroup))
		index = len(layout.Blocks) - 1
	} else {
		layout.Blocks[index] = replaceLegacyCards(layout.Blocks[index], group.DeviceGroup)
	}
	if group.Order != nil {
		// Legacy order refers to group slots only. Standalone integration blocks keep their slots.
		positions, groups := []int{}, []HomeBlock{}
		for i, block := range layout.Blocks {
			if block.Kind == "group" {
				positions = append(positions, i)
				groups = append(groups, block)
			}
		}
		from := slices.IndexFunc(groups, func(block HomeBlock) bool { return block.ID == layout.Blocks[index].ID })
		to := max(0, min(*group.Order, len(groups)-1))
		moved := groups[from]
		groups = slices.Delete(groups, from, from+1)
		groups = slices.Insert(groups, to, moved)
		for i, position := range positions {
			layout.Blocks[position] = groups[i]
		}
	}
	_, err = s.commitHomeLocked(layout, layout.Revision)
	return err
}

func (s *AppConfigCache) importLegacyHomeLocked(groups map[string]*DashboardGroup) error {
	layout, err := s.homeLayoutLocked()
	if err != nil {
		return err
	}
	imported := HomeFromGroups(groups)
	remaining := []HomeBlock{}
	for _, existing := range layout.Blocks {
		if existing.Kind == "card" {
			if existing.Cards[0].Kind != "device" {
				remaining = append(remaining, existing)
			}
			continue
		}
		index := slices.IndexFunc(imported.Blocks, func(block HomeBlock) bool { return block.Name == existing.Name })
		if index >= 0 {
			wanted := groups[existing.Name]
			imported.Blocks[index] = replaceLegacyCards(existing, wanted.DeviceGroup)
			continue
		}
		// Device-only import cannot erase private integration placements from displaced groups.
		for _, card := range existing.Cards {
			if card.Kind != "device" {
				remaining = append(remaining, HomeBlock{ID: legacyHomeID("standalone", card.ID), Kind: "card", Cards: []HomeCard{card}})
			}
		}
	}
	imported.Blocks = append(imported.Blocks, remaining...)
	_, err = s.commitHomeLocked(imported, layout.Revision)
	return err
}
