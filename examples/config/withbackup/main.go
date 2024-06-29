package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/exonlabs/go-utils/pkg/xcfg"
)

var (
	CFGFILE = path.Join(os.TempDir(), "sample_config.json")
	BAKFILE = CFGFILE + ".backup"
)

// mixed Dict and map[string]any definitions
var DEFAULTS = map[string]any{
	"key1": "some value",
	"key2": map[string]any{
		"1": "xxx",
		"2": "yyy",
		"3": "zzz",
	},
}

func _print(msg string, data any) {
	fmt.Println(msg)
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

func main() {
	fmt.Println("\n* using cfg file:", CFGFILE)

	cfg := xcfg.NewJsonConfig(CFGFILE, DEFAULTS)
	cfg.EnableBackup(BAKFILE)
	_print("\n-- initial config:", cfg.Buffer)
	fmt.Println("\n-- saving config")
	if err := cfg.Save(); err != nil {
		panic(err)
	}
	fmt.Println("check master config exist:", cfg.IsFileExist())
	fmt.Println("check backup config exist:", cfg.IsBakFileExist())

	fmt.Println("")

	fmt.Println("-- removing master config")
	os.Remove(CFGFILE)
	fmt.Println("check master config exist:", cfg.IsFileExist())

	fmt.Println("")

	fmt.Println("-- reloading config")
	cfg1 := xcfg.NewJsonConfig(CFGFILE, nil)
	cfg1.EnableBackup("")
	cfg1.Load()
	_print("", cfg1.Buffer)

	cfg.Purge()
	fmt.Println("\n-- config purged")

	fmt.Println()
}
