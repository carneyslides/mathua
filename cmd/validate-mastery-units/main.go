// Command validate_mastery_units checks the draft mastery-unit partition against the corpus.
//
// Failure modes are split by severity, because they are not the same kind of problem:
//
//   - error   — the file is wrong: a concept uncovered or in two units, a dangling reference, a
//     duplicate unit id, membership that has drifted from the rule the file itself declares.
//     These fail the build.
//   - anomaly — the file is internally consistent but the corpus or the rule produces something
//     a human should see: a unit too small to assess, or a derived unit order contradicted by the
//     concept DAG. These are reported and do not fail, because for a draft whose whole purpose
//     is to surface findings, a red build that blocks review is a worse instrument than a
//     loud one. `--strict` promotes them for when the partition stops being a draft.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chuma-beep/mathua/internal/concepts"
)

func main() {
	strict := flag.Bool("strict", false, "treat anomalies as errors too")
	path := flag.String("path", concepts.MasteryUnitsPath, "mastery unit partition to validate")
	flag.Parse()

	if err := run(*strict, *path, "data/concepts", os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
}

func run(strict bool, path, conceptsDir string, stdout, stderr io.Writer) error {
	dag, err := concepts.LoadDir(conceptsDir)
	if err != nil {
		return err
	}
	mu, err := concepts.LoadMasteryUnits(path)
	if err != nil {
		return err
	}

	findings := concepts.ValidateMasteryUnits(dag, mu)
	errs := concepts.Errors(findings)
	anoms := concepts.Anomalies(findings)

	back := concepts.BackwardBandEdges(mu, dag)
	fmt.Fprintf(stdout, "OK: %d concepts partitioned into %d units at band width %d\n",
		mu.ConceptCount, mu.UnitCount, mu.BandWidth)
	for _, f := range anoms {
		fmt.Fprintf(stdout, "  anomaly: %s: %s\n", f.Kind, f.Detail)
	}
	for _, e := range back {
		fmt.Fprintf(stdout, "  backward: %s needs %s (band %d needs band %d)\n",
			e.ConceptID, e.PrereqID, e.Band, e.PrereqBand)
	}

	if len(errs) > 0 {
		for _, f := range errs {
			fmt.Fprintf(stderr, "  error: %s: %s\n", f.Kind, f.Detail)
		}
		return fmt.Errorf("%d partition error(s)", len(errs))
	}
	if strict && len(anoms) > 0 {
		return fmt.Errorf("--strict and %d anomaly/anomalies", len(anoms))
	}
	if len(anoms) > 0 {
		fmt.Fprintf(stdout, "  (%d anomaly/anomalies reported, not failed; pass --strict to promote)\n", len(anoms))
	}
	return nil
}
