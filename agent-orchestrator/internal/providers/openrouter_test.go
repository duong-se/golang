package providers

import (
	"encoding/json"
	"testing"
)

func TestParseToolCalls(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantText      string
		wantToolCount int
		wantToolName  []string
		wantArgs      []map[string]string
	}{
		{
			name:          "no tool call",
			input:         "Hello world",
			wantText:      "Hello world",
			wantToolCount: 0,
		},
		{
			name: "single tool call",
			input: `<tool_call>shell
<arg_key>command</arg_key>
<arg_value>echo hello</arg_value>
<arg_key>description</arg_key>
<arg_value>Create file</arg_value>
</tool_call>`,
			wantText:      "",
			wantToolCount: 1,
			wantToolName: []string{
				"shell",
			},
			wantArgs: []map[string]string{
				{
					"command":     "echo hello",
					"description": "Create file",
				},
			},
		},
		{
			name: "assistant text + tool",
			input: `I'll create the file.

<tool_call>shell
<arg_key>command</arg_key>
<arg_value>echo hello</arg_value>
</tool_call>`,
			wantText:      "I'll create the file.",
			wantToolCount: 1,
			wantToolName: []string{
				"shell",
			},
			wantArgs: []map[string]string{
				{
					"command": "echo hello",
				},
			},
		},
		{
			name: "multiple tool calls",
			input: `<tool_call>shell
<arg_key>command</arg_key>
<arg_value>pwd</arg_value>
</tool_call>

<tool_call>shell
<arg_key>command</arg_key>
<arg_value>ls</arg_value>
</tool_call>`,
			wantText:      "",
			wantToolCount: 2,
			wantToolName: []string{
				"shell",
				"shell",
			},
			wantArgs: []map[string]string{
				{
					"command": "pwd",
				},
				{
					"command": "ls",
				},
			},
		},
		{
			name: "tool without args",
			input: `<tool_call>finish
</tool_call>`,
			wantText:      "",
			wantToolCount: 1,
			wantToolName: []string{
				"finish",
			},
			wantArgs: []map[string]string{
				{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools, text := ParseToolCalls(tt.input)

			if text != tt.wantText {
				t.Fatalf("text mismatch\nwant=%q\ngot=%q", tt.wantText, text)
			}

			if len(tools) != tt.wantToolCount {
				t.Fatalf("tool count mismatch want=%d got=%d",
					tt.wantToolCount,
					len(tools),
				)
			}

			for i := range tools {
				if tools[i].Name != tt.wantToolName[i] {
					t.Fatalf("tool[%d] name want=%q got=%q",
						i,
						tt.wantToolName[i],
						tools[i].Name,
					)
				}

				var got map[string]string

				if err := json.Unmarshal([]byte(tools[i].Arguments), &got); err != nil {
					t.Fatalf("invalid json arguments: %v", err)
				}

				want := tt.wantArgs[i]

				if len(got) != len(want) {
					t.Fatalf("tool[%d] args length mismatch", i)
				}

				for k, v := range want {
					if got[k] != v {
						t.Fatalf("tool[%d] arg %q want=%q got=%q",
							i,
							k,
							v,
							got[k],
						)
					}
				}
			}
		})
	}
}
