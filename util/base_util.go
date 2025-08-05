package util

import "database/sql"

func CompareNullable(a, b any) bool {
	switch aVal := a.(type) {
	case *sql.NullString:
		bVal, ok := b.(*string)
		if !ok {
			return false
		}
		if aVal == nil && bVal == nil {
			return true
		}
		if aVal == nil || bVal == nil {
			return false
		}
		return aVal.String == *bVal

	case *sql.NullInt32:
		bVal, ok := b.(*int32)
		if !ok {
			return false
		}
		if aVal == nil && bVal == nil {
			return true
		}
		if aVal == nil || bVal == nil {
			return false
		}
		return aVal.Int32 == *bVal
	case string:
		bVal, ok := b.(string)
		if !ok {
			return false
		}
		return aVal == bVal
	default:
		return false
	}
}

func IntPtrToNullInt32(i *int) *sql.NullInt32 {
	if i == nil {
		return &sql.NullInt32{Valid: false}
	}
	return &sql.NullInt32{
		Int32: int32(*i),
		Valid: true,
	}
}
