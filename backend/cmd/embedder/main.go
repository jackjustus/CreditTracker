package main

import (
	"fmt"
	"strings"

	"github.com/jackjustus/credittracker/backend/internal/models"
)

func main() {

}

func CoasterDoc(c models.CoasterWithPark) (string, error) {
	var b strings.Builder

	_, err := fmt.Fprintf(&b, "%s is a coaster at %s", c.Coaster().Name(), c.Park().Name())
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
