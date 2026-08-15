package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackjustus/credittracker/backend/internal/data"
	"github.com/jackjustus/credittracker/backend/internal/embed"
	"github.com/jackjustus/credittracker/backend/internal/models"
	"github.com/life4/genesis/slices"
)

func main() {
	ctx := context.Background()
	dao, err := data.NewDAO(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	coasters, err := dao.ListEmbeddingTargets(ctx, embed.Model, 700)
	if err != nil {
		log.Fatal(err)
		return
	}

	docs := slices.Map(coasters, func(c *models.CoasterWithPark) string {
		doc, _ := CoasterDoc(*c)
		return doc
	})

	client, err := embed.NewClient()
	if err != nil {
		log.Fatal(err)
		return
	}

	vecs, err := client.Embed(ctx, docs)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Embed guarantees one vector per doc, so these indexes stay aligned.
	for i, c := range coasters {
		if err := dao.SetCoasterEmbedding(ctx, c.Coaster().ID(), docs[i], embed.Model, vecs[i]); err != nil {
			log.Fatalf("writing %s: %v", c.Coaster().Name(), err)
		}
	}

	fmt.Printf("embedded %d coasters with %s\n", len(coasters), embed.Model)
}

func CoasterDoc(c models.CoasterWithPark) (string, error) {
	var b strings.Builder

	_, err := fmt.Fprintf(&b, "%s is a coaster at %s", c.Coaster().Name(), c.Park().Name())
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
