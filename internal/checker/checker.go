package checker

import (
	"context"
	"errors"
	"fmt"

	"github.com/wutscho/registry-ping/internal/config"
	"github.com/wutscho/registry-ping/internal/notify"
	"github.com/wutscho/registry-ping/internal/registry"
	"github.com/wutscho/registry-ping/internal/state"
)

// scraperFor is a function type to allow testing without a real ScraperRegistry.
type scraperFor interface {
	For(ref registry.ImageRef) (registry.Scraper, error)
}

// Checker orchestrates fetching, comparing, and notifying for a list of images.
type Checker struct {
	scrapers scraperFor
	store    state.StateStore
	notifier notify.Notifier
}

// NewChecker creates a Checker.
func NewChecker(scrapers scraperFor, store state.StateStore, notifier notify.Notifier) *Checker {
	return &Checker{
		scrapers: scrapers,
		store:    store,
		notifier: notifier,
	}
}

// Run checks all images in the config for updates.
// It collects all errors and returns them as a combined error; partial success is allowed.
func (c *Checker) Run(ctx context.Context, images []config.ImageEntry) error {
	var errs []error

	for _, entry := range images {
		if err := c.check(ctx, entry.Ref); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *Checker) check(ctx context.Context, refStr string) error {
	ref, err := registry.ParseImageRef(refStr)
	if err != nil {
		return fmt.Errorf("parse ref %q: %w", refStr, err)
	}

	scraper, err := c.scrapers.For(ref)
	if err != nil {
		return fmt.Errorf("no scraper for %s: %w", ref, err)
	}

	info, err := scraper.Fetch(ctx, ref)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", ref, err)
	}

	key := ref.String()
	prev, found, err := c.store.Load(key)
	if err != nil {
		return fmt.Errorf("load state for %s: %w", ref, err)
	}

	if !found {
		if err := c.notifier.Notify(notify.ChangeEvent{
			Ref:         ref,
			NewPushed:   info.LastPushed,
			IsFirstSeen: true,
		}); err != nil {
			return fmt.Errorf("notify for %s: %w", ref, err)
		}
		if err := c.store.Save(key, state.ImageState{LastPushed: info.LastPushed}); err != nil {
			return fmt.Errorf("save state for %s: %w", ref, err)
		}
		return nil
	}

	if info.LastPushed.After(prev.LastPushed) {
		if err := c.notifier.Notify(notify.ChangeEvent{
			Ref:       ref,
			OldPushed: prev.LastPushed,
			NewPushed: info.LastPushed,
		}); err != nil {
			return fmt.Errorf("notify for %s: %w", ref, err)
		}
		if err := c.store.Save(key, state.ImageState{LastPushed: info.LastPushed}); err != nil {
			return fmt.Errorf("save state for %s: %w", ref, err)
		}
	}
	// else: no change, silent

	return nil
}
