// Package knowledge is the file format of a kdraigo research lab: markdown
// with YAML frontmatter and [[wikilinks]], the shape Obsidian reads. The
// platform keeps a lab's notes, experiments, briefs and handoffs in this form,
// and the MCP writes them to a folder and reads them back (vault sync), so a
// lab opens in Obsidian with its graph.
//
// It holds the format only: parsing and rendering a file, finding its links,
// and naming files. What a lab means is the research service's business.
package knowledge

// Kinds of document in a lab, as the frontmatter's "kind" names them.
const (
	KindBrief      = "brief"      // the lab's one page: goal, constraints, best so far, dead ends, next
	KindNote       = "note"       // an insight, concept, decision or source
	KindExperiment = "experiment" // a hypothesis, its protocol and its verdict
	KindHandoff    = "handoff"    // where an AI session stopped, and what comes next
	KindLesson     = "lesson"     // true across every lab of the user
)

// Kinds lists every kind, in the order a lab shows them.
var Kinds = []string{KindBrief, KindNote, KindExperiment, KindHandoff, KindLesson}

// ValidKind reports whether k is one of Kinds.
func ValidKind(k string) bool {
	for _, kind := range Kinds {
		if k == kind {
			return true
		}
	}
	return false
}
