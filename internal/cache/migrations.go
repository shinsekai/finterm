package cache

import (
	_ "embed"
)

//go:embed migrations/0001_cache_entries.sql
var migration001 string
