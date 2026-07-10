package styles_test

import (
	"fmt"

	"github.com/jarvisfriends/snap/styles"
)

// ExampleActive shows the one accessor every render path uses: the current
// AppStyle, never nil, reflecting the active tint/preset/mode/accessibility
// axes.
func ExampleActive() {
	c := styles.Active()
	fmt.Println(c != nil)
	// Output: true
}
