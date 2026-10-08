// Command plancatalog exports the public subscription offer artifact.
package main

import (
	"encoding/json"
	"flag"
	"github.com/envplane/contracts/domain"
	"log"
	"os"
)

func main() {
	output := flag.String("out", "", "output artifact path")
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
