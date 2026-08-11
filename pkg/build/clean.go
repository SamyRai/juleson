package build

import (
	"context"
	"os"
)

type Cleaner struct {
	binDir string
	files  []string
}

func NewCleaner(binDir string, files []string) *Cleaner {
	return &Cleaner{binDir: binDir, files: files}
}

func (c *Cleaner) Clean(ctx context.Context) error {
	_ = ctx
	if c.binDir != "" {
		if err := os.RemoveAll(c.binDir); err != nil {
			return err
		}
	}
	for _, file := range c.files {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cleaner) CleanAll(ctx context.Context) error {
	if err := c.Clean(ctx); err != nil {
		return err
	}
	if err := c.CleanCache(ctx); err != nil {
		return err
	}
	return c.CleanTestCache(ctx)
}

func (c *Cleaner) CleanCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-cache")
}

func (c *Cleaner) CleanModCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-modcache")
}

func (c *Cleaner) CleanTestCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-testcache")
}
