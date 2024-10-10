package iavl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	idbm "github.com/cosmos/iavl/db"
)

// TODO: make this configurable?
const (
	DefaultCacheSize int = 10000
)

func OpenDB(dir string) (dbm.DB, error) {
	switch {
	case strings.HasSuffix(dir, ".db"):
		dir = dir[:len(dir)-3]
	case strings.HasSuffix(dir, ".db/"):
		dir = dir[:len(dir)-4]
	default:
		return nil, fmt.Errorf("database directory must end with .db")
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// TODO: doesn't work on windows!
	cut := strings.LastIndex(dir, "/")
	if cut == -1 {
		return nil, fmt.Errorf("cannot cut paths on %s", dir)
	}
	name := dir[cut+1:]
	db, err := dbm.NewGoLevelDB(name, dir[:cut], nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type S struct {
	Count  int
	Prefix map[string]int
}

func Stats(db dbm.DB) S {
	count := 0
	prefix := map[string]int{}
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}

	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()[:1]
		prefix[string(key)]++
		count++
	}
	if err := itr.Error(); err != nil {
		panic(err)
	}

	return S{
		Count:  count,
		Prefix: prefix,
	}
}

// ReadTree loads an iavl tree from the directory
// If version is 0, load latest, otherwise, load named version
func ReadTree(db dbm.DB, version int, prefix []byte) (*iavl.MutableTree, error) {
	if len(prefix) == 0 {
		return nil, fmt.Errorf("prefix cannot be empty")
	}
	prefixDB := dbm.NewPrefixDB(db, prefix)
	tree := iavl.NewMutableTree(idbm.NewWrapper(prefixDB), DefaultCacheSize, false, log.NewLogger(os.Stdout))
	_, err := tree.LoadVersion(int64(version))
	return tree, err
}

func ListModules(db dbm.DB) []string {
	modules := make(map[string]struct{})
	it, err := db.Iterator([]byte("s/a"), []byte("s/z"))
	if err != nil {
		panic(err)
	}
	defer it.Close()
	for ; it.Valid(); it.Next() {
		k := it.Key()
		if !strings.HasPrefix(string(k), "s/k:") {
			continue
		}

		split := strings.Split(string(k), "/")
		mod := split[1][2:]
		modules[mod] = struct{}{}
	}

	var keys []string
	for k := range modules {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
