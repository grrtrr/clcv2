package main /*
 * Print details of a single Policy given its policy ID.
 */
import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/grrtrr/clcv2"
	"github.com/grrtrr/exit"
	"github.com/kr/pretty"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]  Policy-ID\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	client, err := clcv2.NewCLIClient()
	if err != nil {
		exit.Fatal(err.Error())
	}

	policy, err := client.SBSgetPolicy(flag.Arg(0))
	if err != nil {
		exit.Fatalf("failed to list SBS policy %s: %s", flag.Arg(0), err)
	}

	pretty.Println(policy)
}
