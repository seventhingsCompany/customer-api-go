package client

import (
	"context"
	"iter"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// defaultPageSize is the page size used by the *All iterators when the caller
// leaves PerPage unset (0). It keeps iteration from making one request per row.
const defaultPageSize = 100

// paginate drives page-walking for an iterator. fetchPage returns the items on
// the given 1-based page. Iteration stops when a page returns fewer than
// perPage items (the last page), when fetchPage errors (the error is yielded
// once), or when the consumer stops ranging. A zero perPage is treated as
// defaultPageSize for the short-page termination check.
func paginate[T any](
	ctx context.Context,
	perPage int,
	fetchPage func(ctx context.Context, page int) ([]T, error),
) iter.Seq2[T, error] {
	if perPage <= 0 {
		perPage = defaultPageSize
	}
	return func(yield func(T, error) bool) {
		for page := 1; ; page++ {
			if err := ctx.Err(); err != nil {
				var zero T
				yield(zero, err)
				return
			}
			items, err := fetchPage(ctx, page)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}
			// A short (or empty) page means we've reached the end.
			if len(items) < perPage {
				return
			}
		}
	}
}

// listOptionsForPage returns a shallow copy of opts (or a fresh one) with Page
// set and PerPage defaulted, so iteration controls paging without mutating the
// caller's value.
func listOptionsForPage(opts *models.ListOptions, page int) *models.ListOptions {
	var o models.ListOptions
	if opts != nil {
		o = *opts
	}
	o.Page = page
	if o.PerPage <= 0 {
		o.PerPage = defaultPageSize
	}
	return &o
}

// ObjectsAll iterates every object across all pages. opts.Page is ignored
// (iteration controls it); opts.PerPage sets the page size and defaults to 100.
func (c *Client) ObjectsAll(ctx context.Context, opts *models.ListOptions) iter.Seq2[models.Fields, error] {
	return paginate(ctx, perPageOf(opts), func(ctx context.Context, page int) ([]models.Fields, error) {
		items, err := c.ObjectsList(ctx, listOptionsForPage(opts, page))
		return asFields(items), err
	})
}

// RoomsAll iterates every room across all pages. See ObjectsAll for paging semantics.
func (c *Client) RoomsAll(ctx context.Context, opts *models.ListOptions) iter.Seq2[models.Fields, error] {
	return paginate(ctx, perPageOf(opts), func(ctx context.Context, page int) ([]models.Fields, error) {
		items, err := c.RoomsList(ctx, listOptionsForPage(opts, page))
		return asFields(items), err
	})
}

// LocationsAll iterates every location across all pages. See ObjectsAll for paging semantics.
func (c *Client) LocationsAll(ctx context.Context, opts *models.ListOptions) iter.Seq2[models.Fields, error] {
	return paginate(ctx, perPageOf(opts), func(ctx context.Context, page int) ([]models.Fields, error) {
		items, err := c.LocationsList(ctx, listOptionsForPage(opts, page))
		return asFields(items), err
	})
}

// CircularityHubItemsAll iterates every circularity-hub item across all pages.
// See ObjectsAll for paging semantics.
func (c *Client) CircularityHubItemsAll(ctx context.Context, opts *models.ListOptions) iter.Seq2[models.Fields, error] {
	return paginate(ctx, perPageOf(opts), func(ctx context.Context, page int) ([]models.Fields, error) {
		items, err := c.CircularityHubItemsList(ctx, listOptionsForPage(opts, page))
		return asFields(items), err
	})
}

// RentalCasesAll iterates every rental case across all pages. See ObjectsAll for paging semantics.
func (c *Client) RentalCasesAll(ctx context.Context, opts *models.ListOptions) iter.Seq2[models.RentalCase, error] {
	return paginate(ctx, perPageOf(opts), func(ctx context.Context, page int) ([]models.RentalCase, error) {
		return c.RentalCasesList(ctx, listOptionsForPage(opts, page))
	})
}

// PersonsAll iterates every person across all pages. opts.Page is ignored;
// opts.PerPage sets the page size and defaults to 100.
func (c *Client) PersonsAll(ctx context.Context, opts *models.PersonListOptions) iter.Seq2[models.Person, error] {
	return paginate(ctx, derefPerPage(personPerPage(opts)), func(ctx context.Context, page int) ([]models.Person, error) {
		o := personOptionsForPage(opts, page)
		resp, err := c.PersonsList(ctx, o)
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	})
}

// UsersAll iterates every user across all pages. opts.Page is ignored;
// opts.PerPage sets the page size and defaults to 100.
func (c *Client) UsersAll(ctx context.Context, opts *models.UserListOptions) iter.Seq2[models.User, error] {
	return paginate(ctx, derefPerPage(userPerPage(opts)), func(ctx context.Context, page int) ([]models.User, error) {
		o := userOptionsForPage(opts, page)
		resp, err := c.UsersList(ctx, o)
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	})
}

// --- helpers ---

func asFields(items []map[string]any) []models.Fields {
	out := make([]models.Fields, len(items))
	for i, m := range items {
		out[i] = models.Fields(m)
	}
	return out
}

func perPageOf(opts *models.ListOptions) int {
	if opts == nil {
		return 0
	}
	return opts.PerPage
}

func derefPerPage(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func personPerPage(opts *models.PersonListOptions) *int {
	if opts == nil {
		return nil
	}
	return opts.PerPage
}

func userPerPage(opts *models.UserListOptions) *int {
	if opts == nil {
		return nil
	}
	return opts.PerPage
}

func personOptionsForPage(opts *models.PersonListOptions, page int) *models.PersonListOptions {
	var o models.PersonListOptions
	if opts != nil {
		o = *opts
	}
	o.Page = &page
	if o.PerPage == nil {
		pp := defaultPageSize
		o.PerPage = &pp
	}
	return &o
}

func userOptionsForPage(opts *models.UserListOptions, page int) *models.UserListOptions {
	var o models.UserListOptions
	if opts != nil {
		o = *opts
	}
	o.Page = &page
	if o.PerPage == nil {
		pp := defaultPageSize
		o.PerPage = &pp
	}
	return &o
}
