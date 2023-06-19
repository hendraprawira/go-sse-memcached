package db

import (
	"os"

	"github.com/bradfitz/gomemcache/memcache"
)

var MC *memcache.Client

func ConnectMemcached() {
	MEMCACHED_DB := os.Getenv("MEMCACHED_DB")
	mc := memcache.New(MEMCACHED_DB)
	MC = mc
}
