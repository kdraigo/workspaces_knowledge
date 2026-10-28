# workspaces_knowledge

The file format of a kdraigo research lab: markdown with YAML frontmatter and
`[[wikilinks]]`, the shape [Obsidian](https://obsidian.md) reads.

The platform keeps a lab's brief, notes, experiments, handoffs and lessons in
this form. The kdraigo MCP writes a lab to a folder and reads it back, so a lab
opens in Obsidian (or any editor) with its links and graph, and a local AI
agent such as Claude Code works on the same knowledge as the one on claude.ai.

```go
import knowledge "github.com/kdraigo/workspaces_knowledge"

doc, err := knowledge.Parse(file)        // frontmatter + body
links := knowledge.Links(doc.Body)       // [[Target#Heading|alias]], code skipped
name := knowledge.Slug("Funding drag")   // "funding-drag": file name and link key
out, err := doc.Render()                 // keys sorted, body untouched
```

| Kind | What it is |
|---|---|
| `brief` | the lab's one page: goal, constraints, best so far, dead ends, next steps |
| `note` | an insight, concept, decision or source |
| `experiment` | a hypothesis, its protocol and its verdict |
| `handoff` | where an AI session stopped, and what comes next |
| `lesson` | true across every lab of the user |

Go 1.25.9. One dependency: `go.yaml.in/yaml/v3`.
