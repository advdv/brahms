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

func draw(t testing.TB, w io.Writer, views map[*Node]View) {
	fmt.Fprintln(w, `digraph {`)
	fmt.Fprintln(w, `layout=neato;`)
	fmt.Fprintln(w, `overlap=scalexy;`)
	fmt.Fprintln(w, `sep="+1";`)

	for self, v := range views {
		fmt.Fprintf(w, "\t"+`"%.8x" [style="filled,solid",label="%s"`, self.Hash().Bytes(), self)
		fmt.Fprintf(w, `,fillcolor="#ffffff"`)
		fmt.Fprintf(w, "]\n")

		for _, n := range v {
			fmt.Fprintf(w, "\t"+`"%.8x" -> "%.8x";`+"\n", self.Hash().Bytes(), n.Hash().Bytes())
		}
	}

	fmt.Fprintln(w, `}`)
}
