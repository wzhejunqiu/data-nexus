package sqlite

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"modernc.org/libc"
	"modernc.org/libc/sys/types"
	sqlite3 "modernc.org/sqlite/lib"
)

const ptrSize = unsafe.Sizeof(uintptr(0))

//nolint:govet // unsafe.Pointer conversions are required for sqlite3 C API interop.
func readByteAt(ptr uintptr) byte {
	if ptr == 0 {
		return 0
	}
	return *(*byte)(unsafe.Pointer(ptr))
}

//nolint:govet // unsafe.Pointer conversions are required for sqlite3 C API interop.
func readUintptrAt(ptr uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(ptr))
}

func (d *Driver) ClassifySQL(ctx context.Context, sqlText string) (model.StatementKind, error) {
	if d.db == nil {
		return "", model.ErrConnectionNotFound("")
	}
	if strings.TrimSpace(sqlText) == "" {
		return "", model.ErrInvalidRequest("sql is required")
	}

	conn, err := d.db.Conn(ctx)
	if err != nil {
		return "", model.ErrInternal(err.Error())
	}
	defer func() { _ = conn.Close() }()

	var kind model.StatementKind
	err = conn.Raw(func(driverConn any) error {
		tls, db, err := extractModerncConn(driverConn)
		if err != nil {
			return model.ErrInternal(err.Error())
		}
		k, err := classifySQLiteStatements(tls, db, sqlText)
		kind = k
		return err
	})
	return kind, err
}

func extractModerncConn(driverConn any) (*libc.TLS, uintptr, error) {
	v := reflect.ValueOf(driverConn)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, 0, fmt.Errorf("unexpected driver conn type %T", driverConn)
	}
	v = v.Elem()
	tlsField := v.FieldByName("tls")
	dbField := v.FieldByName("db")
	if !tlsField.IsValid() || !dbField.IsValid() {
		return nil, 0, fmt.Errorf("driver conn missing tls/db fields")
	}
	if tlsField.Kind() != reflect.Pointer || tlsField.IsNil() {
		return nil, 0, fmt.Errorf("invalid tls on driver conn")
	}
	tls := (*libc.TLS)(tlsField.UnsafePointer())
	return tls, uintptr(dbField.Uint()), nil
}

func classifySQLiteStatements(tls *libc.TLS, db uintptr, sqlText string) (model.StatementKind, error) {
	cstr, err := libc.CString(sqlText)
	if err != nil {
		return "", model.ErrInternal(err.Error())
	}
	defer libc.Xfree(tls, cstr)

	zSQL := cstr
	hasStatement := false

	for {
		zSQL = skipSQLSeparators(zSQL)
		if zSQL == 0 || readByteAt(zSQL) == 0 {
			break
		}

		pstmt, tail, err := prepareStatement(tls, db, zSQL)
		if err != nil {
			return "", err
		}
		if pstmt == 0 {
			zSQL = tail
			continue
		}

		hasStatement = true
		readonly := sqlite3.Xsqlite3_stmt_readonly(tls, pstmt) != 0
		_ = sqlite3.Xsqlite3_finalize(tls, pstmt)
		if !readonly {
			return model.StatementWrite, nil
		}
		zSQL = tail
	}

	if !hasStatement {
		return "", model.ErrInvalidRequest("sql is required")
	}
	return model.StatementQuery, nil
}

func skipSQLSeparators(zSQL uintptr) uintptr {
	for {
		ch := readByteAt(zSQL)
		if ch == 0 {
			return zSQL
		}
		if ch == ';' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			zSQL++
			continue
		}
		return zSQL
	}
}

func prepareStatement(tls *libc.TLS, db, zSQL uintptr) (pstmt, tail uintptr, err error) {
	ppstmt, err := malloc(tls, int(ptrSize))
	if err != nil {
		return 0, 0, model.ErrInternal(err.Error())
	}
	defer libc.Xfree(tls, ppstmt)

	pptail, err := malloc(tls, int(ptrSize))
	if err != nil {
		return 0, 0, model.ErrInternal(err.Error())
	}
	defer libc.Xfree(tls, pptail)

	rc := sqlite3.Xsqlite3_prepare_v2(tls, db, zSQL, -1, ppstmt, pptail)
	if rc != sqlite3.SQLITE_OK {
		return 0, 0, model.ErrSQL(sqliteErrMsg(tls, db))
	}

	pstmt = readUintptrAt(ppstmt)
	tail = readUintptrAt(pptail)
	return pstmt, tail, nil
}

func sqliteErrMsg(tls *libc.TLS, db uintptr) string {
	p := sqlite3.Xsqlite3_errmsg(tls, db)
	if p == 0 {
		return "sqlite error"
	}
	return libc.GoString(p)
}

func malloc(tls *libc.TLS, n int) (uintptr, error) {
	if p := libc.Xmalloc(tls, types.Size_t(n)); p != 0 || n == 0 {
		return p, nil
	}
	return 0, fmt.Errorf("sqlite: cannot allocate %d bytes of memory", n)
}
