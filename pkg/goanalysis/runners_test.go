package goanalysis

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

func Test_buildIssues_suggestedFixes(t *testing.T) {
	// Two files inside the same file set: `a.go` holds the positions [1, 101], `b.go` the positions [102, 202].
	fset := token.NewFileSet()
	fileA := fset.AddFile("a.go", -1, 100)
	fileB := fset.AddFile("b.go", -1, 100)

	analyzer := &analysis.Analyzer{Name: "example"}

	testCases := []struct {
		desc     string
		file     *token.File
		fixes    []analysis.SuggestedFix
		expected []analysis.SuggestedFix
	}{
		{
			desc: "edits inside the file of the diagnostic (control)",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: fileA.Pos(15), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: 10, End: 15, NewText: []byte("atomic.Int64")},
					},
				},
			},
		},
		{
			desc: "edit related to another file",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go and b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "edits split between the file of the diagnostic and another file",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go and b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: fileA.Pos(15), NewText: []byte("atomic.Int64")},
						{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "edits split between another file and the file of the diagnostic",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go and b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
						{Pos: fileA.Pos(10), End: fileA.Pos(15), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "only the suggested fix with an edit related to another file is skipped",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go and b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
					},
				},
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: fileA.Pos(15), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: 10, End: 15, NewText: []byte("atomic.Int64")},
					},
				},
			},
		},
		{
			desc: "edit at the bounds of the file of the diagnostic (control)",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(0), End: fileA.Pos(fileA.Size()), NewText: []byte("package a\n")},
					},
				},
			},
			expected: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: 0, End: token.Pos(fileA.Size()), NewText: []byte("package a\n")},
					},
				},
			},
		},
		{
			desc: "insertion (invalid end) inside the file of the diagnostic (control)",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: token.NoPos, NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: 10, End: 10, NewText: []byte("atomic.Int64")},
					},
				},
			},
		},
		{
			desc: "insertion (invalid end) related to another file",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(10), End: token.NoPos, NewText: []byte(".Load(")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "edit ending inside another file",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: fileB.Pos(15), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "edits related to the second file of the file set (control)",
			file: fileB,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
					},
				},
			},
			expected: []analysis.SuggestedFix{
				{
					Message: "fix b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: 10, End: 15, NewText: []byte(".Load(")},
					},
				},
			},
		},
		{
			desc: "edit related to a file positioned before the file of the diagnostic",
			file: fileB,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(10), End: fileA.Pos(15), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: nil,
		},
		{
			// The edit starts one position past the end of the file of the diagnostic.
			desc: "edit starting at the first position of the next file",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix b.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileB.Pos(0), End: fileB.Pos(5), NewText: []byte(".Load(")},
					},
				},
			},
			expected: nil,
		},
		{
			// The edit ends one position before the start of the file of the diagnostic.
			desc: "edit ending at the last position of the previous file",
			file: fileB,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: fileA.Pos(95), End: fileA.Pos(fileA.Size()), NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "edit without position",
			file: fileA,
			fixes: []analysis.SuggestedFix{
				{
					Message: "fix a.go",
					TextEdits: []analysis.TextEdit{
						{Pos: token.NoPos, End: token.NoPos, NewText: []byte("atomic.Int64")},
					},
				},
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			diag := &Diagnostic{
				Diagnostic: analysis.Diagnostic{
					Pos:            test.file.Pos(10),
					Message:        "message",
					SuggestedFixes: test.fixes,
				},
				Analyzer: analyzer,
				Position: fset.Position(test.file.Pos(10)),
				File:     test.file,
			}

			issues := buildIssues([]*Diagnostic{diag}, func(*Diagnostic) string { return analyzer.Name })

			assert.Len(t, issues, 1)
			assert.Equal(t, test.expected, issues[0].SuggestedFixes)
		})
	}
}

func Test_buildIssues_suggestedFixes_multipleDiagnostics(t *testing.T) {
	// Two files inside the same file set: `a.go` holds the positions [1, 101], `b.go` the positions [102, 202].
	fset := token.NewFileSet()
	fileA := fset.AddFile("a.go", -1, 100)
	fileB := fset.AddFile("b.go", -1, 100)

	analyzer := &analysis.Analyzer{Name: "example"}

	newDiag := func(file *token.File, fixes []analysis.SuggestedFix) *Diagnostic {
		return &Diagnostic{
			Diagnostic: analysis.Diagnostic{
				Pos:            file.Pos(10),
				Message:        "message",
				SuggestedFixes: fixes,
			},
			Analyzer: analyzer,
			Position: fset.Position(file.Pos(10)),
			File:     file,
		}
	}

	// The diagnostic on `a.go` carries an edit related to `b.go`: only that suggested fix is skipped,
	// the suggested fix of the diagnostic on `b.go` is kept.
	diags := []*Diagnostic{
		newDiag(fileA, []analysis.SuggestedFix{
			{
				Message: "fix b.go",
				TextEdits: []analysis.TextEdit{
					{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
				},
			},
		}),
		newDiag(fileB, []analysis.SuggestedFix{
			{
				Message: "fix b.go",
				TextEdits: []analysis.TextEdit{
					{Pos: fileB.Pos(10), End: fileB.Pos(15), NewText: []byte(".Load(")},
				},
			},
		}),
	}

	issues := buildIssues(diags, func(*Diagnostic) string { return analyzer.Name })

	assert.Len(t, issues, 2)
	assert.Nil(t, issues[0].SuggestedFixes)
	assert.Equal(t, []analysis.SuggestedFix{
		{
			Message: "fix b.go",
			TextEdits: []analysis.TextEdit{
				{Pos: 10, End: 15, NewText: []byte(".Load(")},
			},
		},
	}, issues[1].SuggestedFixes)
}
