package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jehiah/nysenateapi"
)

type IndexEntry struct {
	// PrintNo, Session, Chamber, BillType,
	// Title, Summary, LawSection, ActClause
	// LastAction: .Actions[-1]
	PrintNo      string `json:"PrintNo"`
	Session      int    `json:"Session"`
	Chamber      string `json:"Chamber"`
	BillType     string `json:"BillType"`
	Title        string `json:"Title,omitempty"`
	Summary      string `json:"Summary,omitempty"`
	LawSection   string `json:"LawSection,omitempty"`
	ActClause    string `json:"ActClause,omitempty"`
	Status       string `json:"Status,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	SameAs       string `json:"SameAs,omitempty"`
}

func Priority(a, b IndexEntry) IndexEntry {
	for _, status := range []string{"SIGNED_BY_GOV", "VETOED", "DELIVERED_TO_GOV", "PASSED_ASSEMBLY", "PASSED_SENATE"} {
		if a.Status == status {
			return a
		}
		if b.Status == status {
			return b
		}
	}
	if a.LastModified != "" {
		return a
	}
	if b.LastModified != "" {
		return b
	}
	return a
}

func NewIndexEntry(bill nysenateapi.Bill) IndexEntry {
	var lastModifiedDate string
	if len(bill.Actions) > 0 {
		lastModifiedDate = bill.Actions[len(bill.Actions)-1].Date.String()
	}
	same, _, _ := strings.Cut(bill.SameAsPrintNo, "-")
	return IndexEntry{
		PrintNo:      bill.PrintNo,
		Session:      bill.Session,
		Chamber:      bill.Chamber,
		BillType:     bill.BillType,
		Title:        bill.Title,
		Summary:      StripParts(bill.Summary),
		LawSection:   bill.LawSection,
		ActClause:    StripParts(bill.ActClause),
		LastModified: lastModifiedDate,
		Status:       bill.Status,
		SameAs:       same,
	}
}

func buildIndex(legislationDir, outputFile string) error {
	os.Mkdir(filepath.Dir(outputFile), 0777)
	var index []IndexEntry
	same := make(map[string]IndexEntry)
	// walk all json files in legislationDir, read them, and build an index of the following format:
	err := filepath.WalkDir(legislationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		var bill nysenateapi.Bill
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&bill)
		if err != nil {
			return err
		}
		i := NewIndexEntry(bill)
		if existing, ok := same[i.SameAs]; ok {
			i = Priority(existing, i)
			// remove the existing entry from the index
			if i.PrintNo != existing.PrintNo {
				for j, entry := range index {
					if entry.PrintNo == existing.PrintNo {
						log.Printf("removing duplicate entry for %s (same as %s)", existing.PrintNo, i.PrintNo)
						index = append(index[:j], index[j+1:]...)
						break
					}
				}
				same[i.PrintNo] = i
				index = append(index, i)
			}
		} else {
			same[i.PrintNo] = i
			index = append(index, i)
		}

		// read the json file, extract the relevant fields, and add them to the index
		return nil
	})
	if err != nil {
		return err
	}

	// write index
	log.Printf("writing index with %d entries to %s", len(index), outputFile)
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(index)
}

func params(base, output, billType string, year int) (string, string) {
	return filepath.Join(base, billType, strconv.Itoa(year)), filepath.Join(output, fmt.Sprintf("%s_%d.json", billType, year))
}

func main() {
	legislationDir := flag.String("legislation_dir", "../../../ny_legislation", "Path to checkout of ny_legislation")
	outputDir := flag.String("output_dir", "../../build", "Path to output directory for index files")
	year := flag.Int("year", 2025, "Year of legislation to build index for")
	flag.Parse()

	if err := buildIndex(params(*legislationDir, *outputDir, "bills", *year)); err != nil {
		slog.Error("Failed to build bills index", "error", err)
		os.Exit(1)
	}
	if err := buildIndex(params(*legislationDir, *outputDir, "resolutions", *year)); err != nil {
		slog.Error("Failed to build laws index", "error", err)
		os.Exit(1)
	}

}
