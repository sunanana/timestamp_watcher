package timestamp_watcher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

//TODO: 構造体名簡略化しすぎ、簡略化やっていいのは変数名だけ。
type M map[string]*F
type F struct {
	ModTime *time.Time
}

var (
	m1, m2      M
	targetIsDir bool
	ignore      []string
)

type WatchConfig struct {
	IntervalSec       uint
	Target            string
	Ignore            []string
	IsPrintChange     bool
	PrintChangeWriter io.Writer
}

func init() {
	m1, m2 = M{}, M{}
}

func Watch(t *WatchConfig) (chan struct{}, error) {
	if err := valid(t); err != nil {
		return nil, err
	}

	fInfo, _ := os.Stat(t.Target)
	targetIsDir = fInfo.IsDir()
	ignore = t.Ignore

	ch := make(chan struct{})
	watch(t, ch)
	return ch, nil
}

func watch(t *WatchConfig, ch chan struct{}) {
	recursiveDig(lsLa(t.Target), m1)

	go func() {
		for range time.NewTicker(time.Duration(t.IntervalSec) * time.Second).C {
			m2 = M{}
			recursiveDig(lsLa(t.Target), m2)

			if !reflect.DeepEqual(m1, m2) {
				ch <- struct{}{}
				if t.IsPrintChange {
					go printDiff(t.PrintChangeWriter, m1, m2)
				}

				m1 = M{}
				for k, v := range m2 {
					m1[k] = v
				}
			}
		}
	}()
}

func recursiveDig(ret []string, m M) {
	l, i := len(ret), -1
	for {
	skipIngnore:
		if i++; l == i {
			break
		}

		f := ret[i]
		for _, v := range ignore {
			if f == v {
				goto skipIngnore
			}
		}

		fInfo, err := os.Lstat(f)
		if err != nil {
			panic(err)
		}

		if fInfo.IsDir() {
			t := fInfo.ModTime()
			m[f] = &F{ModTime: &t}

			a := lsLa(f)
			if len(a) != 0 {
				recursiveDig(a, m)
			}
		} else {
			// TODO: symlinkの先にあるディレクトリは循環参照無限ループのためチェックしていない,要解決.
			t := fInfo.ModTime()
			m[f] = &F{ModTime: &t}
		}
	}
}

func lsLa(n string) []string {
	var t string
	if targetIsDir {
		t = n + "/*"
	} else {
		t = n
	}

	a, err := filepath.Glob(t)
	if err != nil {
		panic(err)
	}
	return a
}

// TODO: `mv old_dir new_dif` をDelete&Addとして表示する. `Move: old_dir -> new_dif` として表示したい.
func printDiff(w io.Writer, t1, t2 M) {
	if w == nil {
		w = os.Stdout
	}

	tmp1, tmp2 := M{}, M{}
	for k, v := range t1 {
		tmp1[k] = v
	}
	for k, v := range t2 {
		tmp2[k] = v
	}

	// Print Delete
	for k := range tmp1 {
		if _, exists := tmp2[k]; exists {
			continue
		}
		fmt.Fprint(w, fmt.Sprintf("Delete: %s\n", k))
		delete(tmp1, k)
	}

	// Print Add
	for k := range tmp2 {
		if _, exists := tmp1[k]; exists {
			continue
		}
		fmt.Fprint(w, fmt.Sprintf("Add: %s\n", k))
		delete(tmp2, k)
	}

	// Print Modify
	for k, v := range tmp2 {
		if !reflect.DeepEqual(tmp1[k], v) {
			fmt.Fprint(w, fmt.Sprintf("Modify: %s\n", k))
		}
	}
}

func valid(t *WatchConfig) error {
	errs := []string{}
	if t.IntervalSec == 0 {
		errs = append(errs, "interval time should not be 0.")
	}
	if t.Target == "" {
		errs = append(errs, "target should not be blank.")
	} else {
		if _, err := os.Stat(t.Target); err != nil {
			errs = append(errs, fmt.Sprintf("not found target. Target: %s", t.Target))
		}
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\t"))
	}
	return nil
}
