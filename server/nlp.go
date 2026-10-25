package main

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var skillDictionary = map[string][]string{
	"Go":                  {"golang", "go lang"},
	"Python":              {"python3", "py"},
	"Java":                {},
	"JavaScript":          {"js", "java script"},
	"TypeScript":          {"ts"},
	"C++":                 {"cpp", "c plus plus"},
	"C":                   {},
	"SQL":                 {"mysql", "postgres", "postgresql", "sqlite"},
	"React":               {"reactjs", "react.js"},
	"Node.js":             {"node", "nodejs"},
	"Angular":             {"angularjs"},
	"Django":              {},
	"Spring Boot":         {"spring"},
	"Android":             {"android development"},
	"iOS":                 {"swift", "ios development"},
	"Flutter":             {"dart"},
	"HTML":                {},
	"CSS":                 {"tailwind", "bootstrap"},
	"Docker":              {"containers"},
	"Kubernetes":          {"k8s"},
	"AWS":                 {"amazon web services"},
	"Azure":               {},
	"Google Cloud":        {"gcp"},
	"Terraform":           {},
	"Linux":               {"unix"},
	"Git":                 {"github", "version control"},
	"Machine Learning":    {"ml", "deep learning", "neural networks"},
	"Data Science":        {"data analysis", "analytics"},
	"NLP":                 {"natural language processing"},
	"Computer Vision":     {"opencv"},
	"Data Structures":     {"dsa", "algorithms"},
	"System Design":       {"distributed systems", "scalability"},
	"Cyber Security":      {"security", "penetration testing", "infosec"},
	"Blockchain":          {"solidity", "smart contracts", "web3"},
	"Embedded Systems":    {"microcontroller", "arduino", "raspberry pi"},
	"VLSI":                {"verilog", "vhdl"},
	"Signal Processing":   {"dsp"},
	"Robotics":            {"ros"},
	"AutoCAD":             {"auto cad"},
	"SolidWorks":          {"solid works"},
	"CATIA":               {},
	"Structural Analysis": {"staad", "structural design"},
	"Surveying":           {"total station"},
	"Project Management":  {"pmp", "agile", "scrum"},
	"Thermodynamics":      {},
	"Manufacturing":       {"production", "cnc"},
	"Electric Vehicles":   {"ev", "battery systems"},
	"Product Design":      {"ui design", "ux", "figma"},
	"Technical Writing":   {"documentation"},
	"Public Speaking":     {"presentation"},
	"Teaching":            {"mentoring", "training"},
}

var skillPatterns = buildSkillPatterns()

type skillPattern struct {
	canonical string
	pattern   *regexp.Regexp
}

func buildSkillPatterns() []skillPattern {
	patterns := make([]skillPattern, 0, len(skillDictionary))

	for canonical, aliases := range skillDictionary {
		terms := append([]string{canonical}, aliases...)

		escaped := make([]string, len(terms))
		for i, term := range terms {
			escaped[i] = regexp.QuoteMeta(term)
		}

		expression := `(?i)(^|[^\w+#.])(` + strings.Join(escaped, "|") + `)($|[^\w+#])`
		patterns = append(patterns, skillPattern{canonical, regexp.MustCompile(expression)})
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].canonical < patterns[j].canonical
	})
	return patterns
}

func extractSkills(text string) []string {
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	found := []string{}
	for _, entry := range skillPatterns {
		if entry.pattern.MatchString(text) {
			found = append(found, entry.canonical)
		}
	}
	return found
}

type extractRequest struct {
	Text string `json:"text"`
}

func handleExtractSkills(w http.ResponseWriter, r *http.Request, user *User) {
	var body extractRequest
	if !readJSON(w, r, &body) {
		return
	}

	profile, err := loadProfile(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load your profile")
		return
	}

	text := body.Text
	if strings.TrimSpace(text) == "" {
		text = profile.Bio
	}

	existing := make(map[string]bool)
	for _, skill := range splitList(profile.Skills) {
		existing[skill] = true
	}

	found := extractSkills(text)

	fresh := []string{}
	already := []string{}
	for _, skill := range found {
		if existing[strings.ToLower(skill)] {
			already = append(already, skill)
		} else {
			fresh = append(fresh, skill)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"found":           found,
		"new":             fresh,
		"already":         already,
		"dictionary_size": len(skillDictionary),
	})
}
