package barnacle

import (
	"fmt"
	"io"

	"github.com/fordaj/barnacle/tools/lucide"
)

// ListIcons writes the sorted list of icons embedded in the binary (one
// per line) to w, framed by a header and a trailing attribution block that
// credits Lucide — name, link and license — as required when redistributing
// the set. It is the implementation behind `barnacle list-icons`.
func ListIcons(w io.Writer) error {
	names, err := lucide.IconNames()
	if err != nil {
		return err
	}
	version, err := lucide.Version()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "barnacle — %d icons from %s v%s (%s)\n\n", len(names), lucide.Name, version, lucide.Homepage)
	for _, name := range names {
		fmt.Fprintln(w, name)
	}
	fmt.Fprintf(w, "\n%s\n", lucide.Attribution())
	return nil
}
