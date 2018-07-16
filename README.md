# Example
```go
package main

import (
	"fmt"

	tw "./timestamp_watcher"
)

func main() {
	ch := tw.Watch(&tw.WatchConfig{
		IntervalSec:   3,
		Target:        ".",
		Ignore:        []string{".git", ".idea"},
		IsPrintChange: true,
	})

	for range ch {
		fmt.Println("Timestamp update.")
	}
}
```

# 動機
Docker for Macのvolumesマウントがホスト側のファイル更新を検知してくれないので作った。