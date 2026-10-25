package main

import (
	"strings"
	"testing"
)

func TestExtractSkills(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		want  []string
		avoid []string
	}{
		{
			name: "plain mentions",
			text: "I build backend services in Go and Python.",
			want: []string{"Go", "Python"},
		},
		{
			name: "aliases resolve to canonical names",
			text: "Worked with golang, k8s and gcp for three years.",
			want: []string{"Go", "Kubernetes", "Google Cloud"},
		},
		{
			name: "multi word phrases",
			text: "My focus is natural language processing and deep learning.",
			want: []string{"NLP", "Machine Learning"},
		},
		{
			name:  "go is not found inside other words",
			text:  "I had to google the algorithm and forgot it again.",
			avoid: []string{"Go"},
		},
		{
			name:  "python is not found inside pythonic",
			text:  "We kept the code pythonic.",
			avoid: []string{"Python"},
		},
		{
			name: "punctuation does not block a match",
			text: "Stack: React, Node.js, SQL.",
			want: []string{"React", "Node.js", "SQL"},
		},
		{
			name: "c++ is matched despite symbols",
			text: "Competitive programming in C++.",
			want: []string{"C++"},
		},
		{
			name:  "c++ mention does not also report bare C",
			text:  "Competitive programming in C++.",
			avoid: []string{"C"},
		},
		{
			name: "civil engineering vocabulary",
			text: "Ten years of AutoCAD drawings and structural design on highways.",
			want: []string{"AutoCAD", "Structural Analysis"},
		},
		{
			name: "empty text finds nothing",
			text: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			found := extractSkills(c.text)
			index := make(map[string]bool)
			for _, skill := range found {
				index[skill] = true
			}

			for _, want := range c.want {
				if !index[want] {
					t.Errorf("expected %q in %v", want, found)
				}
			}
			for _, avoid := range c.avoid {
				if index[avoid] {
					t.Errorf("did not expect %q in %v", avoid, found)
				}
			}
			if c.text == "" && len(found) != 0 {
				t.Errorf("expected nothing, got %v", found)
			}
			_ = strings.TrimSpace
		})
	}
}
