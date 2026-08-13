package main

import (
	"context"
	"fmt"
)

func run(ctx context.Context, cfg *Config) error {
	fmt.Println("Hello World!")
	return nil
}