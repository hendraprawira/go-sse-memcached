package db

import (
	"os"

	"github.com/bradfitz/gomemcache/memcache"
)

var MC *memcache.Client

func ConnectMemchaced() {
	MEMCHACED_DB := os.Getenv("MEMCHACED_DB")
	mc := memcache.New(MEMCHACED_DB)
	MC = mc
}
