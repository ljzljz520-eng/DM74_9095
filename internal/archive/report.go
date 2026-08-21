package archive

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type ArchiveIndex struct {
	Entries []domain.ArchiveEntry
	ByBatch map[string]string
}

type ArchiveQuery struct {
	BatchID    string
	ArchivedBy string
	Class      string
}

func (s *Service) BuildIndex() (ArchiveIndex, error) {
	entries, err := s.store.ListArchives()
	if err != nil {
		return ArchiveIndex{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ArchivedOn == entries[j].ArchivedOn {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].ArchivedOn < entries[j].ArchivedOn
	})
	index := ArchiveIndex{Entries: entries, ByBatch: make(map[string]string, len(entries))}
	for _, entry := range entries {
		index.ByBatch[entry.BatchID] = entry.ID
	}
	return index, nil
}

func (i ArchiveIndex) Find(query ArchiveQuery) []domain.ArchiveEntry {
	result := make([]domain.ArchiveEntry, 0)
	for _, entry := range i.Entries {
		if query.BatchID != "" && entry.BatchID != query.BatchID {
			continue
		}
		if query.ArchivedBy != "" && !strings.EqualFold(entry.ArchivedBy, query.ArchivedBy) {
			continue
		}
		if query.Class != "" && Classify(entry) != query.Class {
			continue
		}
		result = append(result, entry)
	}
	return result
}

func (i ArchiveIndex) Counts() map[string]int {
	counts := make(map[string]int)
	for _, entry := range i.Entries {
		counts[Classify(entry)]++
	}
	return counts
}

func (i ArchiveIndex) Summary() string {
	counts := i.Counts()
	return fmt.Sprintf("archives=%d complete=%d partial=%d rejected=%d", len(i.Entries), counts["complete"], counts["partial"], counts["rejected"])
}

func (s *Service) ValidateIndex(index ArchiveIndex) error {
	seen := make(map[string]bool, len(index.Entries))
	for _, entry := range index.Entries {
		if entry.ID == "" || entry.BatchID == "" {
			return fmt.Errorf("archive identity is incomplete")
		}
		if seen[entry.ID] {
			return fmt.Errorf("duplicate archive %s", entry.ID)
		}
		seen[entry.ID] = true
		if mapped := index.ByBatch[entry.BatchID]; mapped != entry.ID {
			return fmt.Errorf("batch %s has inconsistent index", entry.BatchID)
		}
		if len(entry.RejectedIDs) > len(entry.ItemIDs) {
			return fmt.Errorf("archive %s has too many rejected items", entry.ID)
		}
	}
	return nil
}

func (s *Service) ArchiveDigest() (string, error) {
	index, err := s.BuildIndex()
	if err != nil {
		return "", err
	}
	if err := s.ValidateIndex(index); err != nil {
		return "", err
	}
	lines := make([]string, 0, len(index.Entries)+1)
	lines = append(lines, index.Summary())
	for _, entry := range index.Entries {
		lines = append(lines, fmt.Sprintf("%s %s %s by=%s", entry.ID, entry.BatchID, Classify(entry), entry.ArchivedBy))
	}
	return strings.Join(lines, "\n"), nil
}
