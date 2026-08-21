package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"theatre39/internal/domain"
	"theatre39/internal/store"
	"theatre39/internal/workflow39"
)

func main() {
	dbPath := flag.String("db", "theatre39.db", "path to the theatre reservation database")
	demo := flag.Bool("demo", false, "run a deterministic registration example")
	flag.Parse()
	if err := run(*dbPath, *demo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path string, demo bool) error {
	if path == "" {
		return fmt.Errorf("database path cannot be empty")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Clean(path)
	}
	db, err := store.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	if !demo {
		counts, countErr := db.SnapshotCounts()
		if countErr != nil {
			return countErr
		}
		fmt.Printf("community theatre reservation store ready: %d batches, %d items\n", counts["batches"], counts["items"])
		return nil
	}
	coordinator := workflow39.New(db)
	batch, err := coordinator.RegisterAndSubmit(domain.BatchInput{ID: "demo-batch", Theatre: "Community Hall", Performance: "friday-19", CreatedBy: "front-desk", SubmittedNote: "demo intake"}, []domain.ItemInput{{ID: "demo-item-1", Patron: "Demo Patron", SeatCode: "C-10", RequestedClass: "standard"}})
	if err != nil {
		return err
	}
	report, err := coordinator.ReviewAndNotify(batch.ID, "Demo Reviewer", "desk")
	if err != nil {
		return err
	}
	fmt.Println(report.String())
	return nil
}
