package wails

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// RuntimePort abstracts Wails desktop runtime calls for testability.
type RuntimePort interface {
	WindowSetTitle(ctx context.Context, title string)
	MessageDialog(ctx context.Context, opts wailsruntime.MessageDialogOptions) (string, error)
	EventsEmit(ctx context.Context, event string, data ...any)
	OpenFileDialog(ctx context.Context, opts wailsruntime.OpenDialogOptions) (string, error)
}

type wailsRuntime struct{}

func (wailsRuntime) WindowSetTitle(ctx context.Context, title string) {
	wailsruntime.WindowSetTitle(ctx, title)
}

func (wailsRuntime) MessageDialog(ctx context.Context, opts wailsruntime.MessageDialogOptions) (string, error) {
	return wailsruntime.MessageDialog(ctx, opts)
}

func (wailsRuntime) EventsEmit(ctx context.Context, event string, data ...any) {
	wailsruntime.EventsEmit(ctx, event, data...)
}

func (wailsRuntime) OpenFileDialog(ctx context.Context, opts wailsruntime.OpenDialogOptions) (string, error) {
	return wailsruntime.OpenFileDialog(ctx, opts)
}
