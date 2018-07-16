# Example
```go
package main

import (
	"fmt"

	tw "github.com/sunanana/timestamp_watcher"
)

func main() {
	ch, err := tw.Watch(&tw.WatchConfig{
		IntervalSec:   3,
		Target:        ".",
		Ignore:        []string{".git", ".idea"},
		IsPrintChange: true,
	})
	if err != nil {
		panic(err)
	}

	for range ch {
		fmt.Println("Timestamp update.")
	}
}
```

# 動機
Docker for Macのvolumesマウントがホスト側のファイル更新を検知してくれないので作った。