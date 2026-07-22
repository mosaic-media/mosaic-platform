// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package screens

import (
	"context"
	"sync"

	sdui "github.com/mosaic-media/sdui/sdui"
	"github.com/mosaic-media/sdui/ui"

	"github.com/mosaic-media/platform/internal/platform/app"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

const (
	homeMaxRows     = 6
	homeMaxRowItems = 20
	// homeUpNextItems bounds the "Up next" filmstrip docked on the hero floor —
	// the items neighbouring the featured one, drawn from the same first catalog.
	homeUpNextItems = 8
	// homeHeroSlides bounds the hero carousel: the top item of each of the first
	// few non-empty catalogs, auto-advancing behind the content sheet.
	homeHeroSlides = 5
)

// homeScreen is the default landing surface: a hero over rows of the enabled
// modules' catalogs (Cinemeta's Popular Movies/Series, etc. — ADR 0028's virtual
// plane, browsed not materialised). Each row is a carousel of cards that open a
// detail; the hero is the first catalog's first item, enriched with its backdrop
// and logo. Browsing is a read, so nothing here writes.
func (s *Service) homeScreen(ctx context.Context, caller v1.Caller) (sdui.Node, error) {
	cats, err := s.content.ListModuleCatalogs(ctx, app.ListModuleCatalogsQuery{Caller: caller})
	if err != nil {
		return nil, err
	}
	if len(cats.Catalogs) == 0 {
		return ui.Screen(ui.EmptyState(emptyIconCollections,
			"Nothing here yet — add an addon in Settings to browse content")).Build(), nil
	}

	// Render at most homeMaxRows rows. Each row's items are a remote round-trip,
	// so fetch them concurrently rather than serially — the landing page pays one
	// round-trip instead of a sum. We fetch only the catalogs we render (the first
	// homeMaxRows), bounding remote load to the visible rows; a catalog beyond that
	// is not fetched, and one that errors simply drops its row.
	catalogs := cats.Catalogs
	if len(catalogs) > homeMaxRows {
		catalogs = catalogs[:homeMaxRows]
	}
	itemsByCatalog := make([]app.ListCatalogItemsResult, len(catalogs))
	var wg sync.WaitGroup
	for i, c := range catalogs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// A downed catalog leaves its slot empty, which the assembly below skips
			// — the same effect as the serial code's continue-on-error.
			items, err := s.content.ListCatalogItems(ctx, app.ListCatalogItemsQuery{
				Caller: caller, ModuleID: c.ModuleID, CatalogID: c.Catalog.ID, NativeType: c.Catalog.NativeType,
			})
			if err == nil {
				itemsByCatalog[i] = items
			}
		}()
	}
	wg.Wait()

	// Assemble the page as a widget tree. The featured banner comes from the
	// first non-empty catalog's first item (one further round-trip to enrich it),
	// spanning the Screen's full-bleed slot with an "Up next" filmstrip of its
	// neighbours docked on its floor; then a carousel row per non-empty catalog.
	rows := make([]ui.El, 0, len(catalogs)+1)
	type heroPick struct {
		item   v1.CatalogItem
		kicker string
	}
	var picks []heroPick
	var upNext ui.El
	for i, c := range catalogs {
		items := itemsByCatalog[i].Items
		if len(items) == 0 {
			continue
		}
		// The hero carousel takes the top item of each of the first few catalogs.
		if len(picks) < homeHeroSlides {
			picks = append(picks, heroPick{item: items[0], kicker: c.Catalog.Name})
		}
		if upNext == nil {
			// "Trending now" — the items neighbouring the first featured one — leads
			// the library as a rail of glass MediaTiles, the showcase row for the
			// acrylic material (the edge light tracks the hero art across the row).
			// Library rows below stay plain PosterCards.
			upCards := make([]ui.El, 0, homeUpNextItems)
			for j := 1; j < len(items) && j <= homeUpNextItems; j++ {
				it := items[j]
				upCards = append(upCards, s.mediaTile(it.Ref, it.Title, it.Year, it.Poster, it.InLibrary))
			}
			if len(upCards) > 0 {
				upNext = ui.Section("Trending now", ui.Carousel(upCards...))
			}
		}
		cards := make([]ui.El, 0, homeMaxRowItems)
		for j, it := range items {
			if j >= homeMaxRowItems {
				break
			}
			cards = append(cards, s.contentCard(it.Ref, it.Title, it.Year, it.Poster, it.InLibrary))
		}
		rows = append(rows, ui.Section(c.Catalog.Name, ui.Carousel(cards...)))
	}
	if len(rows) == 0 {
		return ui.Screen(ui.EmptyState(emptyIconCollections,
			"Nothing to show yet — try adding an addon in Settings")).Build(), nil
	}

	// Enrich the featured picks into hero banners concurrently — each is a further
	// metadata round-trip (backdrop/logo). Order is preserved; a pick whose
	// enrichment fails drops out.
	slides := make([]ui.El, len(picks))
	var hg sync.WaitGroup
	for i, p := range picks {
		hg.Add(1)
		go func() {
			defer hg.Done()
			if h := s.heroFromItem(ctx, caller, p.item, p.kicker); h != nil {
				slides[i] = h
			}
		}()
	}
	hg.Wait()
	heroSlides := make([]ui.El, 0, len(slides))
	for _, h := range slides {
		if h != nil {
			heroSlides = append(heroSlides, h)
		}
	}

	// The home is a cinematic backdrop the content rides over. A Rotator auto-
	// advances the hero slides (mostly artwork — no pills/overview/buttons) and is
	// `sticky`, so it stays put while the library, carried on a glass "sheet",
	// scrolls UP over it: the sheet's acrylic top edge catches the active hero's
	// light on the way past. Both live in the Screen's edge-to-edge `bleed` slot
	// (the sheet owns its own gutter), so the padded body collapses ($childCount 0).
	// When enrichment failed for every pick there's no hero; the sheet stands alone.
	sheetEls := make([]ui.El, 0, len(rows)+2)
	sheetEls = append(sheetEls, ui.Prop("style", map[string]any{
		"glass": true, "bg": "bg", "radius": "xl",
		"direction": "column", "gap": 8,
		"px": "gutter", "pt": 8, "pb": 9,
		"position": "relative", "z": "raised",
	}))
	if upNext != nil {
		sheetEls = append(sheetEls, upNext)
	}
	sheetEls = append(sheetEls, rows...)
	sheet := ui.Component("Box", sheetEls...)

	bleed := make([]ui.El, 0, 2)
	if len(heroSlides) > 0 {
		rotEls := make([]ui.El, 0, len(heroSlides)+1)
		rotEls = append(rotEls, ui.Prop("interval", 6000))
		rotEls = append(rotEls, heroSlides...)
		bleed = append(bleed, ui.Component("Rotator", rotEls...))
	}
	bleed = append(bleed, sheet)
	return ui.Screen(ui.Slot("bleed", bleed...)).Build(), nil
}

// heroFromItem builds the home's featured banner from a catalog item, enriching
// it with the backdrop, logo and overview its lightweight card lacks (ADR 0034).
// It is full-bleed and tagged with the catalog it leads (the `kicker`). A
// metadata fetch that fails just yields no hero (nil) rather than failing the
// home screen.
func (s *Service) heroFromItem(ctx context.Context, caller v1.Caller, it v1.CatalogItem, kicker string) *ui.Element {
	prev, err := s.content.PreviewContent(ctx, app.PreviewContentQuery{Caller: caller, Ref: it.Ref})
	if err != nil {
		return nil
	}
	m := prev.Metadata
	title := m.Title
	if title == "" {
		title = it.Title
	}

	// The home hero is artwork first: just the catalog kicker + the title (or logo)
	// over the backdrop — no overview, no meta pills, no buttons. That "detail page"
	// chrome belongs on the detail screen; here the poster is the hero.
	return ui.Hero(title,
		ui.Prop("variant", "feature"),
		ui.When(kicker != "", ui.Prop("kicker", kicker)),
		ui.Backdrop(s.art(m.Backdrop)),
		ui.When(m.Logo != "", ui.Logo(s.art(m.Logo))),
	)
}
