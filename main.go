package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

var counts = make(map[string]int)

// emitFull adds `weight` to the count of the full pattern formed by taking the
// group's shared labels (rep) and substituting kLabel at the varying position k.
func emitFull(rep []string, k int, kLabel string, weight int) {
	var sb strings.Builder
	for i := range rep {
		if i == k {
			sb.WriteString(kLabel)
		} else {
			sb.WriteString(rep[i])
		}
		if i != len(rep)-1 {
			sb.WriteByte('.')
		}
	}
	counts[sb.String()] += weight
}

// buildDashWild joins parts with "-", wildcarding the part at index p, e.g.
// (["m33","out"], 1) -> "m33-{replace_this}".
func buildDashWild(parts []string, p int) string {
	var sb strings.Builder
	for i := range parts {
		if i > 0 {
			sb.WriteByte('-')
		}
		if i == p {
			sb.WriteString("{replace_this}")
		} else {
			sb.WriteString(parts[i])
		}
	}
	return sb.String()
}

type dashedVal struct {
	parts []string
	freq  int
}

// subAgg accumulates, within one dash-part sub-group, the distinct values seen at
// the varying part together with their frequencies.
type subAgg struct {
	rep  []string
	vals map[string]int
}

func (s *subAgg) add(v string, f int) {
	if s.vals == nil {
		s.vals = make(map[string]int)
	}
	s.vals[v] += f
}

// pairWeight is the number of ordered pairs of distinct values, S^2 - sum(f^2).
func (s *subAgg) pairWeight() int {
	S, Q := 0, 0
	for _, f := range s.vals {
		S += f
		Q += f * f
	}
	return S*S - Q
}

// processGroup emits every pattern for a set of domains that share all labels
// except the one at position k. `freq` maps the varying label value to how many
// domains carried it. All weights are computed in closed form — no pairwise loop
// — which is what makes a group with tens of thousands of distinct values cheap.
//
// Every ordered pair of distinct values falls into exactly one bucket:
//   - SKIP  (emit nothing): only at k==0, when the pair is not both-dashed with
//     equal dash-part counts (mirrors the original's `break` at position 0).
//   - FINE  (dash-decomposed wildcard): both dashed, equal part counts, differing
//     in exactly one dash-part.
//   - COARSE (whole-label wildcard): everything else that isn't skipped.
func processGroup(rep []string, k int, freq map[string]int) {
	T, sumSq := 0, 0
	for _, f := range freq {
		T += f
		sumSq += f * f
	}
	allPairs := T*T - sumSq // ordered pairs of distinct values
	if allPairs == 0 {
		return
	}

	// Bucket the dashed values by dash-part count; only these can be FINE, and at
	// k==0 only these escape the SKIP rule.
	countClass := make(map[int][]dashedVal)
	for v, f := range freq {
		if strings.Contains(v, "-") {
			parts := strings.Split(v, "-")
			countClass[len(parts)] = append(countClass[len(parts)], dashedVal{parts, f})
		}
	}

	// SKIP weight: at k==0, pairs that are NOT both-dashed-with-equal-part-counts
	// are dropped (the original breaks and emits nothing).
	skip := 0
	if k == 0 {
		nonskip := 0
		for _, dv := range countClass {
			s, q := 0, 0
			for _, x := range dv {
				s += x.freq
				q += x.freq * x.freq
			}
			nonskip += s*s - q
		}
		skip = allPairs - nonskip
	}

	// FINE weight: within each part-count class, for each dash-part position p,
	// sub-group by the *other* parts. Every pair inside a sub-group differs in
	// exactly that one part, so its whole weight maps to a single fine pattern —
	// again closed form.
	fineTotal := 0
	for c, dv := range countClass {
		if c < 2 {
			continue
		}
		for p := 0; p < c; p++ {
			sub := make(map[string]*subAgg)
			var kb strings.Builder
			for _, x := range dv {
				kb.Reset()
				for q := 0; q < c; q++ {
					if q == p {
						continue
					}
					kb.WriteString(x.parts[q])
					kb.WriteByte(0)
				}
				key := kb.String()
				sg := sub[key]
				if sg == nil {
					sg = &subAgg{rep: x.parts}
					sub[key] = sg
				}
				sg.add(x.parts[p], x.freq)
			}
			for _, sg := range sub {
				w := sg.pairWeight()
				if w == 0 {
					continue
				}
				fineTotal += w
				emitFull(rep, k, buildDashWild(sg.rep, p), w)
			}
		}
	}

	// COARSE gets whatever is left: total minus skipped minus fine.
	if coarse := allPairs - skip - fineTotal; coarse > 0 {
		emitFull(rep, k, "{replace_this}", coarse)
	}
}

func logic(domainMap map[int][][]string) {
	maxLength2process := 14

	for L := 2; L <= maxLength2process; L++ {
		domains := domainMap[L]
		if len(domains) < 2 {
			continue
		}

		// For each label position, bucket domains by all the *other* labels.
		// Domains in one bucket differ only at position k — exactly the single
		// {replace_this} case — so unrelated pairs are never touched.
		for k := 0; k < L; k++ {
			groups := make(map[string][]string) // key -> freq is built per group below
			reps := make(map[string][]string)
			var keyb strings.Builder
			for _, d := range domains {
				keyb.Reset()
				for i := 0; i < L; i++ {
					if i == k {
						continue
					}
					keyb.WriteString(d[i])
					keyb.WriteByte(0) // NUL separator: never appears in a label
				}
				key := keyb.String()
				if _, ok := reps[key]; !ok {
					reps[key] = d
				}
				groups[key] = append(groups[key], d[k])
			}

			for key, vals := range groups {
				if len(vals) < 2 {
					continue
				}
				freq := make(map[string]int, len(vals))
				for _, v := range vals {
					freq[v]++
				}
				if len(freq) < 2 {
					continue // only one distinct value -> no differing pair
				}
				processGroup(reps[key], k, freq)
			}
		}
	}
}

func main() {
	domainMap := make(map[int][][]string)

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		splittedDomain := strings.Split(domain, ".")
		domainLength := len(splittedDomain)
		domainMap[domainLength] = append(domainMap[domainLength], splittedDomain)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	logic(domainMap)

	// Emit "<count> <pattern>" sorted ascending by count, replacing the
	// external `sort | uniq -c | sort -n` pipeline.
	patterns := make([]string, 0, len(counts))
	for p := range counts {
		patterns = append(patterns, p)
	}
	sort.Slice(patterns, func(a, b int) bool {
		if counts[patterns[a]] != counts[patterns[b]] {
			return counts[patterns[a]] < counts[patterns[b]]
		}
		return patterns[a] < patterns[b]
	})

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, p := range patterns {
		fmt.Fprintf(w, "%d %s\n", counts[p], p)
	}
}
