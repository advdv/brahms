package brahms

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/advanderveer/go-test"
)

func drawPNG(t *testing.T, buf io.Reader, name string) {
	f, err := os.Create(name)
	test.Ok(t, err)
	defer f.Close()

	cmd := exec.Command("neato", "-Tpng")
	cmd.Stdin = buf
	cmd.Stdout = f
	test.Ok(t, cmd.Run())
}

func draw(t testing.TB, w io.Writer, cores []*Core) {
	fmt.Fprintln(w, `digraph {`)
	fmt.Fprintln(w, `layout=neato;`)
	fmt.Fprintln(w, `overlap=scalexy;`)
	fmt.Fprintln(w, `sep="+1";`)

	for _, c := range cores {
		fmt.Fprintf(w, "\t"+`"%.6x" [style="filled,solid",label="%.6x"`, c.ID(), c.ID())
		fmt.Fprintf(w, `,fillcolor="#ffffff"`)
		fmt.Fprintf(w, "]\n")

		for id := range c.view {
			fmt.Fprintf(w, "\t"+`"%.6x" -> "%.6x";`+"\n", c.ID(), id)
		}
	}

	fmt.Fprintln(w, `}`)
}
