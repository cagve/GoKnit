package main 

import (
	"fmt"
	"strings"
)

func ToString[T any](items []T) string {
	var parts []string
	for _, item  := range items {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	return strings.Join(parts, ", ")

}
