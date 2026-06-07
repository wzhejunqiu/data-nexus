package sqlite

import (
	"context"
	"fmt"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

type attachState struct {
	mu       sync.Mutex
	attached []model.AttachedDatabase
	readOnly bool
}

func (d *Driver) attachState() *attachState {
	if d.attachedDBs == nil {
		d.attachedDBs = &attachState{readOnly: d.readOnly}
	}
	return d.attachedDBs
}

func (d *Driver) Attach(ctx context.Context, filePath, alias string) error {
	if !sqlutil.IsSafeQuotedIdentifier(alias) {
		return model.ErrInvalidRequest("invalid attach alias")
	}
	if alias == "main" || alias == "temp" {
		return model.ErrInvalidRequest("reserved alias")
	}
	st := d.attachState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, a := range st.attached {
		if a.Alias == alias {
			return model.ErrInvalidRequest(fmt.Sprintf("alias already attached: %s", alias))
		}
	}
	attachPath := filePath
	if d.readOnly {
		attachPath = "file:" + filePath + "?mode=ro"
	}
	_, err := d.db.ExecContext(ctx, fmt.Sprintf("ATTACH DATABASE ? AS %q", alias), attachPath)
	if err != nil {
		return model.ErrSQL(err.Error())
	}
	st.attached = append(st.attached, model.AttachedDatabase{Alias: alias, FilePath: filePath})
	return nil
}

func (d *Driver) Detach(ctx context.Context, alias string) error {
	if !sqlutil.IsSafeQuotedIdentifier(alias) {
		return model.ErrInvalidRequest("invalid detach alias")
	}
	st := d.attachState()
	st.mu.Lock()
	defer st.mu.Unlock()
	found := false
	var next []model.AttachedDatabase
	for _, a := range st.attached {
		if a.Alias == alias {
			found = true
			continue
		}
		next = append(next, a)
	}
	if !found {
		return model.ErrInvalidRequest(fmt.Sprintf("alias not attached: %s", alias))
	}
	if _, err := d.db.ExecContext(ctx, fmt.Sprintf("DETACH %q", alias)); err != nil {
		return model.ErrSQL(err.Error())
	}
	st.attached = next
	return nil
}

func (d *Driver) ListAttached(ctx context.Context) ([]model.AttachedDatabase, error) {
	st := d.attachState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]model.AttachedDatabase, len(st.attached))
	copy(out, st.attached)
	return out, nil
}

func (d *Driver) detachAll() {
	st := d.attachState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, a := range st.attached {
		_, _ = d.db.Exec(fmt.Sprintf("DETACH %q", a.Alias))
	}
	st.attached = nil
}
