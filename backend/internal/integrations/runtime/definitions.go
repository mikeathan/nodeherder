package runtime

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"

	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func (s *Service) loadDefinitions() error {
	if s.shipped != nil {
		if e := fs.WalkDir(s.shipped, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".yaml") {
				return nil
			}
			key := strings.TrimSuffix(path, ".yaml")
			existing := false
			if e := s.repo.View(func(tx *repository.IntegrationTx) error {
				var record model.DefinitionRecord
				var e error
				existing, e = tx.Get("definitions", key, &record)
				return e
			}); e != nil {
				return e
			}
			if existing {
				return nil
			}
			_, _, e := s.files.Read(key)
			if os.IsNotExist(e) {
				b, e := s.shippedDefinition(path)
				if e != nil {
					return e
				}
				e = s.files.Write(key, string(b), "")
				if errors.Is(e, repository.ErrDefinitionConflict) {
					return nil
				}
				return e
			}
			return e
		}); e != nil {
			return e
		}
	}
	keys, e := s.files.List()
	if e != nil {
		return e
	}
	seen := map[string]bool{}
	for _, key := range keys {
		seen[key] = true
		if e = s.reloadDefinition(key, "", nil, true); e != nil {
			var f *model.Failure
			if !errors.As(e, &f) {
				return e
			}
		}
	}
	return s.repo.Update(func(tx *repository.IntegrationTx) error {
		missing := []model.DefinitionRecord{}
		if e := tx.Each("definitions", func(_ string, b []byte) error {
			var r model.DefinitionRecord
			if e := repository.DecodeIntegrationRecord(b, &r); e != nil {
				return e
			}
			if !seen[r.Key] {
				r.DiskError = "The provider file is missing. Restore it on the host and Reload."
				missing = append(missing, r)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, r := range missing {
			if e := tx.Put("definitions", r.Key, r); e != nil {
				return e
			}
		}
		return nil
	})
}

// Definition reads host source and active metadata without contacting a provider. Missing or
// unreadable files retain the last active source so the editor can offer explicit recovery.
func (s *Service) Definition(key string) (DefinitionDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out DefinitionDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		var r model.DefinitionRecord
		ok, e := tx.Get("definitions", key, &r)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Provider not found.")
		}
		source, hash, e := s.files.Read(key)
		readError := e
		if e != nil {
			source = r.YAML
			hash = ""
		}
		out = DefinitionDTO{
			Key:         key,
			YAML:        source,
			ContentHash: hash,
			ActiveHash:  r.Hash,
			Revision:    r.Revision,
			ETag:        definitions.StrongETag(hash),
			Status:      "active",
			DiskError:   r.DiskError,
		}
		if c, e := s.compiler.CompileYAML(source); e == nil {
			out.Document = c.Definition
		} else {
			out.Status = "invalid"
		}
		if hash != "" && hash != r.Hash {
			out.Status = "host_changed"
		}
		if r.PendingHash != "" {
			out.Status = "saved_not_applied"
		}
		if r.DiskError != "" {
			out.Status = "disk_error"
		}
		if readError != nil {
			out.Status = "disk_error"
			out.DiskError = "The host file could not be read. The last active source is shown."
			if os.IsNotExist(readError) {
				out.Status = "missing"
				out.DiskError = "The host file is missing. Recreate it from this draft or restore the last active file."
			}
		}
		if b, e := s.shippedDefinition(key + ".yaml"); e == nil {
			out.ShippedYAML = string(b)
			out.UpdateAvailable = r.SeedHash != "" && definitions.ContentHash(b) != r.SeedHash
		}
		return nil
	})
	return out, e
}

// ValidateDefinition checks source or Form JSON without writing files or sending requests.
// It also reports the origin/credential revisions requiring explicit review before activation.
func (s *Service) ValidateDefinition(key, source string, document json.RawMessage) (ValidationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.validateDefinition(key, source, document)
}
func (s *Service) validateDefinition(key, source string, document json.RawMessage) (ValidationDTO, error) {
	var out ValidationDTO
	var c *definitions.Compiled
	var e error
	if document != nil {
		c, e = s.compiler.CompileDocument(document)
	} else {
		c, e = s.compiler.CompileYAML(source)
	}
	if e != nil {
		return out, e
	}
	if key != "" && definitions.DefinitionKey(c.Definition) != key {
		return out, model.InvalidDefinition("id", "The type and provider ID must match the selected provider file.")
	}
	out = ValidationDTO{YAML: c.YAML, Document: c.Definition, Hash: c.Hash, Origin: c.Origin, Affected: []CredentialReview{}}
	if s.repo == nil {
		return out, nil
	}
	e = s.repo.View(func(tx *repository.IntegrationTx) error {
		var r model.DefinitionRecord
		ok, e := tx.Get("definitions", key, &r)
		if e != nil {
			return e
		}
		if ok && r.YAML != "" {
			old, e := s.compiler.CompileYAML(r.YAML)
			if e != nil {
				return e
			}
			out.PreviousOrigin = old.Origin
		}

		return tx.Each("instances", func(_ string, b []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
				return e
			}
			if i.DefinitionKey != key {
				return nil
			}
			changed := out.PreviousOrigin != "" && out.PreviousOrigin != out.Origin
			activeCredential := credential(c.Definition, i.Credentials)
			credentialMoved := activeCredential != nil && activeCredential.Origin != out.Origin
			if !changed && i.Origin == out.Origin && !credentialMoved {
				return nil
			}
			if out.PreviousOrigin == out.Origin {
				out.PreviousOrigin = i.Origin
				if credentialMoved {
					out.PreviousOrigin = activeCredential.Origin
				}
			}
			rev := uint64(0)
			if cred := credential(c.Definition, i.Credentials); cred != nil {
				rev = cred.Revision
			}
			out.Affected = append(out.Affected, CredentialReview{CredentialSlot: c.Definition.Auth.SecretSlot, InstanceID: i.ID, Revision: i.Revision, CredentialRevision: rev})
			return nil
		})
	})
	return out, e
}

// SaveDefinition journals the new source before atomic host replacement and runtime activation.
func (s *Service) SaveDefinition(key, source, expected string, create bool, review *DestinationReview) (DefinitionDTO, error) {
	s.mu.Lock()
	if e := s.ready(); e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	validation, e := s.validateDefinition(key, source, nil)
	if e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	if e = reviewDestination(validation, review); e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	_, hash, e := s.files.Read(key)
	if create {
		if e == nil {
			s.mu.Unlock()
			return DefinitionDTO{}, model.Fail("revision_conflict", "This provider file already exists.")
		}
		if !os.IsNotExist(e) {
			s.mu.Unlock()
			return DefinitionDTO{}, e
		}
		hash = ""
	} else {
		if e != nil {
			s.mu.Unlock()
			return DefinitionDTO{}, model.Fail("config_directory_unwritable", "The provider file could not be read.")
		}
		if e = checkRevision(expected, definitions.StrongETag(hash)); e != nil {
			s.mu.Unlock()
			return DefinitionDTO{}, e
		}
	}
	var record model.DefinitionRecord
	e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		_, e := tx.Get("definitions", key, &record)
		if e != nil {
			return e
		}
		record.Key = key
		record.PendingYAML = source
		record.PendingHash = validation.Hash
		return tx.Put("definitions", key, record)
	})
	if e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	if e = s.files.Write(key, source, hash); e != nil {
		if !errors.Is(e, repository.ErrDefinitionConflict) {
			_, writtenHash, readErr := s.files.Read(key)
			if readErr == nil && writtenHash == validation.Hash {
				s.mu.Unlock()
				return DefinitionDTO{}, &model.Failure{
					Code:        "saved_not_applied",
					Message:     "The host file contains these changes, but saving/applying could not be fully confirmed. The previous version remains active; use Reload to recover.",
					ContentHash: writtenHash,
					ActiveHash:  record.Hash,
					RetryAction: "reload",
				}
			}
		}
		_ = s.repo.Update(func(tx *repository.IntegrationTx) error {
			record.PendingYAML = ""
			record.PendingHash = ""
			return tx.Put("definitions", key, record)
		})
		s.mu.Unlock()
		if errors.Is(e, repository.ErrDefinitionConflict) {
			return DefinitionDTO{}, model.Fail("revision_conflict", "The host file changed. Reload it before saving.")
		}
		return DefinitionDTO{}, model.Fail("config_directory_unwritable", "The provider file could not be saved. Check the configuration mount.")
	}
	e = s.activateDefinition(key, source, validation.Hash, review)
	s.mu.Unlock()
	if e != nil {
		return DefinitionDTO{}, &model.Failure{
			Code:        "saved_not_applied",
			Message:     "The file was saved but could not be activated. The previous definition remains active; use Reload to recover.",
			ContentHash: validation.Hash,
			ActiveHash:  record.Hash,
			RetryAction: "reload",
		}
	}
	return s.Definition(key)
}

// reviewDestination requires consent outside the editable file. Configuration text is
// descriptive and cannot grant itself permission to send saved inputs/credentials elsewhere.
func reviewDestination(v ValidationDTO, r *DestinationReview) error {
	if v.PreviousOrigin != "" && v.Origin != v.PreviousOrigin {
		if r == nil || r.Origin != v.Origin {
			return model.Fail("destination_review_required", "Review the new provider destination before saving.")
		}
		for _, given := range r.Reassociate {
			found := false
			for _, actual := range v.Affected {
				if given == actual {
					found = true
				}
			}
			if !found {
				return model.Fail("revision_conflict", "A integration changed after destination review. Validate again.")
			}
		}
	}
	return nil
}

// activateDefinition commits the revision and invalidates obsolete work together. The
// host file has already been written/journaled; a failure leaves that mismatch observable
// for explicit retry/restart recovery rather than pretending the active source changed.
func (s *Service) activateDefinition(key, source, hash string, review *DestinationReview) error {
	c, e := s.compiler.CompileYAML(source)
	if e != nil {
		return e
	}
	return s.repo.Update(func(tx *repository.IntegrationTx) error {
		var record model.DefinitionRecord
		_, e := tx.Get("definitions", key, &record)
		if e != nil {
			return e
		}
		if record.Hash == hash && record.PendingHash == "" && record.DiskError == "" && review == nil {
			return nil
		}
		record.Key = key
		if seed, e := s.shippedDefinition(key + ".yaml"); e == nil {
			seedHash := definitions.ContentHash(seed)
			if record.SeedHash == "" || hash == seedHash {
				record.SeedHash = seedHash
			}
		}
		oldHash := record.Hash
		if oldHash != hash {
			record.PreviousYAML = record.YAML
			record.PreviousExpiresAt = s.clock.Now().Add(7 * 24 * time.Hour)
		}
		record.YAML = source
		record.Hash = hash
		record.PendingHash = ""
		record.PendingYAML = ""
		record.DiskError = ""
		record.Revision++
		if e = tx.Put("definitions", key, record); e != nil {
			return e
		}
		if oldHash == hash && review == nil {
			return nil
		}
		updates := []model.Instance{}
		if e = tx.Each("instances", func(_ string, b []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
				return e
			}
			if i.DefinitionKey == key {
				updates = append(updates, i)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, i := range updates {
			oldRevision := i.Revision
			oldOperationID := i.OperationID
			i.Revision++
			i.Generation++
			i.ETag = ""
			i.CacheScope = ""
			currentCredential := credential(c.Definition, i.Credentials)
			if i.Origin != c.Origin || currentCredential != nil && currentCredential.Origin != c.Origin {
				reassociate := false
				if review != nil {
					for _, grant := range review.Reassociate {
						cred := credential(c.Definition, i.Credentials)
						revision := uint64(0)
						if cred != nil {
							revision = cred.Revision
						}
						if grant.InstanceID == i.ID && grant.Revision == oldRevision && grant.CredentialRevision == revision && grant.CredentialSlot == c.Definition.Auth.SecretSlot {
							reassociate = true
						}
					}
				}
				if reassociate {
					i.Origin = c.Origin
					i.LastError = nil
					if cred := credential(c.Definition, i.Credentials); cred != nil {
						cred.Origin = c.Origin
						cred.Revision++
						i.Credentials[c.Definition.Auth.SecretSlot] = *cred
					}
				} else {
					i.LastError = model.Fail("reconnect_required", "The provider destination changed. Review and reconnect before sending saved inputs.")
				}
				if e = s.terminateOperation(tx, i.OperationID, model.Fail("reconnect_required", "The destination changed. Start a reviewed setup.")); e != nil {
					return e
				}
				i.OperationID = ""
			} else if i.OperationID != "" {
				var op model.Operation
				ok, e := tx.Get("operations", i.OperationID, &op)
				if e != nil {
					return e
				}
				if ok && op.Status != "done" && op.Status != "failed" {
					op.Generation = i.Generation
					op.Status = "retry_required"
					op.RetryAt = s.clock.Now().Add(5 * time.Second)
					if e = tx.Put("operations", op.ID, op); e != nil {
						return e
					}
				}
			}
			if cancel := s.running[oldOperationID]; cancel != nil {
				cancel()
			}
			if e = tx.Put("instances", i.ID, i); e != nil {
				return e
			}
		}
		return nil
	})
}

// ReloadDefinition validates and activates the current host file under an optimistic hash check.
func (s *Service) ReloadDefinition(key, expected string, review *DestinationReview) (DefinitionDTO, error) {
	s.mu.Lock()
	if e := s.ready(); e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	e := s.reloadDefinition(key, expected, review, false)
	s.mu.Unlock()
	if e != nil {
		return DefinitionDTO{}, e
	}
	return s.Definition(key)
}
func (s *Service) reloadDefinition(key, expected string, review *DestinationReview, startup bool) error {
	source, hash, e := s.files.Read(key)
	if e != nil {
		return model.Fail("config_directory_unwritable", "The provider file is missing or unreadable. Last saved results are retained.")
	}
	if !startup {
		if e = checkRevision(expected, definitions.StrongETag(hash)); e != nil {
			return e
		}
	}
	validation, e := s.validateDefinition(key, source, nil)
	if e != nil {
		saveErr := s.repo.Update(func(tx *repository.IntegrationTx) error {
			var r model.DefinitionRecord
			if _, err := tx.Get("definitions", key, &r); err != nil {
				return err
			}
			r.Key = key
			r.DiskError = "The host file is invalid. Repair it in Configure operations; the last good revision is retained."
			return tx.Put("definitions", key, r)
		})
		if saveErr != nil {
			return saveErr
		}
		return e
	}
	if !startup {
		if e = reviewDestination(validation, review); e != nil {
			return e
		}
	}
	return s.activateDefinition(key, source, hash, review)
}

// RestoreDefinition writes the retained previous good source as a new revision; it does not
// roll back private connection data or bypass changed-destination review.
func (s *Service) RestoreDefinition(key, expected string, review *DestinationReview) (DefinitionDTO, error) {
	s.mu.Lock()
	var previous string
	e := s.ready()
	if e == nil {
		e = s.repo.View(func(tx *repository.IntegrationTx) error {
			var r model.DefinitionRecord
			ok, e := tx.Get("definitions", key, &r)
			if e != nil {
				return e
			}
			if !ok || r.PreviousYAML == "" {
				return model.Fail("not_found", "No previous provider revision is available.")
			}
			previous = r.PreviousYAML
			return nil
		})
	}
	s.mu.Unlock()
	if e != nil {
		return DefinitionDTO{}, e
	}
	return s.SaveDefinition(key, previous, expected, false, review)
}

// DeleteDefinition removes an unused host definition. Live instances/setup references block deletion.
func (s *Service) DeleteDefinition(key, expected string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return e
	}
	_, hash, e := s.files.Read(key)
	if e != nil {
		return model.Fail("not_found", "Provider file not found.")
	}
	if e = checkRevision(expected, definitions.StrongETag(hash)); e != nil {
		return e
	}
	if e = s.repo.View(func(tx *repository.IntegrationTx) error {
		for _, bucket := range []string{"instances", "setups", "operations"} {
			if e := tx.Each(bucket, func(_ string, b []byte) error {
				var record map[string]any
				if e := repository.DecodeIntegrationRecord(b, &record); e != nil {
					return e
				}
				if record["definition_key"] == key {
					return model.Fail("definition_in_use", "Remove this provider's integrations and setups before deleting its definition.")
				}
				return nil
			}); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		return e
	}
	if e = s.files.Delete(key, hash); e != nil {
		return e
	}
	return s.repo.Update(func(tx *repository.IntegrationTx) error { return tx.Delete("definitions", key) })
}

// ReloadDefinitions discovers host files without fetching provider data or granting credentials.
func (s *Service) ReloadDefinitions() (CatalogDTO, error) {
	s.mu.Lock()
	if e := s.ready(); e != nil {
		s.mu.Unlock()
		return CatalogDTO{}, e
	}
	e := s.loadDefinitions()
	s.mu.Unlock()
	if e != nil {
		return CatalogDTO{}, e
	}
	return s.Catalog()
}

// AcknowledgeShipped keeps a reviewed custom host file without repeatedly announcing the same seed.
func (s *Service) AcknowledgeShipped(key, expected string) (DefinitionDTO, error) {
	s.mu.Lock()
	if e := s.ready(); e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	_, hash, e := s.files.Read(key)
	if e == nil {
		e = checkRevision(expected, definitions.StrongETag(hash))
	}
	if e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, e
	}
	seed, e := s.shippedDefinition(key + ".yaml")
	if e != nil {
		s.mu.Unlock()
		return DefinitionDTO{}, model.Fail("not_found", "No shipped example exists for this provider.")
	}
	e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var r model.DefinitionRecord
		ok, e := tx.Get("definitions", key, &r)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Provider not found.")
		}
		r.SeedHash = definitions.ContentHash(seed)
		return tx.Put("definitions", key, r)
	})
	s.mu.Unlock()
	if e != nil {
		return DefinitionDTO{}, e
	}
	return s.Definition(key)
}

// shippedDefinition treats absent bundled defaults like a missing optional file.
func (s *Service) shippedDefinition(path string) ([]byte, error) {
	if s.shipped == nil {
		return nil, fs.ErrNotExist
	}
	return fs.ReadFile(s.shipped, path)
}
