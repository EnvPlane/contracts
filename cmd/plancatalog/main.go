// Command plancatalog exports the public subscription offer artifact.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/envplane/contracts/domain"
	"log"
	"os"
)

func main() {
	output := flag.String("out", "", "output artifact path")
	check := flag.String("check", "", "check an existing artifact for canonical drift")
	flag.Parse()
	catalog := domain.CommercialPlanCatalog().Deterministic()
	if err := catalog.Validate(); err != nil {
		log.Fatal(err)
	}
	encoded, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *check != "" {
		existing, err := os.ReadFile(*check)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(existing, encoded) {
			log.Fatal("pricing artifact differs from canonical CommercialPlanCatalog")
		}
		return
	}
	if *output == "" {
		if _, err := os.Stdout.Write(encoded); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := os.WriteFile(*output, encoded, 0600); err != nil {
		log.Fatal(err)
	}
}
