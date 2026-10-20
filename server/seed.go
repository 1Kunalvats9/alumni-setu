package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

const seedPassword = "password123"

type seedPerson struct {
	Name     string
	Email    string
	Role     string
	Verified bool
	Profile  Profile
}

var seedPeople = []seedPerson{
	{"Department Admin", "admin@rtu.ac.in", "admin", true, Profile{
		College: "Technical Education Dept, Rajasthan", Location: "Jaipur",
	}},

	{"Arjun Mehta", "arjun@alumni.ac.in", "alumni", true, Profile{
		College: "MNIT Jaipur", Branch: "Computer Science", GraduationYear: 2016,
		Company: "Google", JobTitle: "Senior Software Engineer", Industry: "Software",
		Location: "Bengaluru", Skills: "Go, Kubernetes, Distributed Systems, Python",
		Interests: "Open Source, Mentoring, System Design",
		Bio:       "Backend engineer. Happy to review resumes and do mock interviews.",
		LinkedIn:  "https://linkedin.com/in/arjunmehta", OpenToMentor: true,
	}},
	{"Priya Singh", "priya@alumni.ac.in", "alumni", true, Profile{
		College: "RTU Kota", Branch: "Electronics", GraduationYear: 2018,
		Company: "Infosys", JobTitle: "Data Scientist", Industry: "Consulting",
		Location: "Pune", Skills: "Python, Machine Learning, SQL, Statistics",
		Interests: "Data Science, Teaching, Women in Tech",
		Bio:       "Moved from core electronics to data science. Ask me how.",
		LinkedIn:  "https://linkedin.com/in/priyasingh", OpenToMentor: true,
	}},
	{"Vikram Rao", "vikram@alumni.ac.in", "alumni", true, Profile{
		College: "MNIT Jaipur", Branch: "Civil Engineering", GraduationYear: 2014,
		Company: "L&T Construction", JobTitle: "Project Manager", Industry: "Construction",
		Location: "Jaipur", Skills: "AutoCAD, Project Management, Site Planning",
		Interests:    "Infrastructure, Public Projects",
		Bio:          "Fifteen highway and metro projects across Rajasthan.",
		OpenToMentor: true,
	}},
	{"Neha Agarwal", "neha@alumni.ac.in", "alumni", true, Profile{
		College: "JECRC Jaipur", Branch: "Computer Science", GraduationYear: 2019,
		Company: "Razorpay", JobTitle: "Frontend Engineer", Industry: "Fintech",
		Location: "Bengaluru", Skills: "JavaScript, React, TypeScript, CSS",
		Interests:    "Product Design, Open Source, Mentoring",
		Bio:          "I build payment dashboards. Ask me about frontend interviews.",
		OpenToMentor: true,
	}},
	{"Rahul Jain", "rahul@alumni.ac.in", "alumni", false, Profile{
		College: "RTU Kota", Branch: "Mechanical Engineering", GraduationYear: 2017,
		Company: "Tata Motors", JobTitle: "Design Engineer", Industry: "Automotive",
		Location: "Pune", Skills: "SolidWorks, CATIA, Thermodynamics",
		Interests: "Electric Vehicles, Manufacturing",
		Bio:       "Working on EV powertrain design.",
	}},
	{"Sana Qureshi", "sana@alumni.ac.in", "alumni", true, Profile{
		College: "MNIT Jaipur", Branch: "Computer Science", GraduationYear: 2020,
		Company: "Microsoft", JobTitle: "Cloud Engineer", Industry: "Software",
		Location: "Hyderabad", Skills: "Azure, Go, Kubernetes, Terraform",
		Interests: "Cloud, Open Source, Mentoring",
		Bio:       "Cloud infrastructure. Big fan of Go.", OpenToMentor: true,
	}},

	{"Riya Sharma", "riya@student.ac.in", "student", false, Profile{
		College: "MNIT Jaipur", Branch: "Computer Science", GraduationYear: 2027,
		Location: "Jaipur", Skills: "Python, Go, DSA",
		Interests: "Open Source, Software, System Design",
		Bio:       "Third year CSE. Looking for a backend internship.",
	}},
	{"Aditya Verma", "aditya@student.ac.in", "student", false, Profile{
		College: "RTU Kota", Branch: "Electronics", GraduationYear: 2026,
		Location: "Kota", Skills: "C, Embedded Systems, Python",
		Interests: "Data Science, Robotics",
		Bio:       "Final year. Interested in moving into data roles.",
	}},
	{"Meera Gupta", "meera@student.ac.in", "student", false, Profile{
		College: "JECRC Jaipur", Branch: "Computer Science", GraduationYear: 2027,
		Location: "Jaipur", Skills: "JavaScript, React, CSS",
		Interests: "Product Design, Fintech, Open Source",
		Bio:       "I like building interfaces. Learning TypeScript now.",
	}},
	{"Karan Yadav", "karan@student.ac.in", "student", false, Profile{
		College: "MNIT Jaipur", Branch: "Civil Engineering", GraduationYear: 2026,
		Location: "Jaipur", Skills: "AutoCAD, Surveying",
		Interests: "Infrastructure, Public Projects",
		Bio:       "Want to work on public infrastructure in Rajasthan.",
	}},
}

type seedEvent struct {
	Event    Event
	Creator  string
	Attendee []string
}

var seedEvents = []seedEvent{
	{Event{Title: "Careers in Cloud Engineering", Kind: "webinar",
		Description: "Live Q&A on cloud roles, certifications and interview prep.",
		StartsAt:    "2026-11-15T10:00:00Z", Link: "https://meet.example/cloud"},
		"sana@alumni.ac.in", []string{"riya@student.ac.in", "aditya@student.ac.in"}},

	{Event{Title: "Rajasthan Alumni Meetup 2026", Kind: "meetup",
		Description: "Annual in-person meetup for alumni and final year students.",
		StartsAt:    "2026-12-05T05:30:00Z", Location: "MNIT Jaipur Auditorium"},
		"arjun@alumni.ac.in", []string{"riya@student.ac.in", "meera@student.ac.in", "karan@student.ac.in"}},

	{Event{Title: "Panel: Breaking into Data Science", Kind: "panel",
		Description: "Three alumni share how they switched from core branches to data roles.",
		StartsAt:    "2026-11-28T12:00:00Z", Link: "https://meet.example/data"},
		"priya@alumni.ac.in", []string{"aditya@student.ac.in"}},
}

type seedPost struct {
	Post     Post
	Author   string
	Comments []seedComment
}

type seedComment struct {
	Author string
	Body   string
}

var seedPosts = []seedPost{
	{Post{Title: "How should I prepare for backend interviews?",
		Body:     "I am in third year CSE. Which topics matter most and how deep should I go?",
		Category: "career"},
		"riya@student.ac.in",
		[]seedComment{
			{"arjun@alumni.ac.in", "Pick one language and go deep. Then DBMS, OS and networks. Build two projects you can defend line by line."},
			{"sana@alumni.ac.in", "Agreed. Also practise explaining your projects out loud, that is half the interview."},
		}},

	{Post{Title: "Is a masters worth it before joining industry?",
		Body:     "Confused between placements and preparing for GATE. Would like to hear both sides.",
		Category: "academic"},
		"aditya@student.ac.in",
		[]seedComment{
			{"priya@alumni.ac.in", "I joined industry first and did a part time masters later. Work experience made the coursework far more useful."},
		}},

	{Post{Title: "Frontend internship openings for 2027 batch",
		Body:     "Sharing a few openings I came across. Ask here if you want a referral.",
		Category: "placement"},
		"neha@alumni.ac.in",
		[]seedComment{
			{"meera@student.ac.in", "This is really helpful, thank you. Sending you a mentorship request."},
		}},
}

type seedJob struct {
	Job    Job
	Poster string
}

var seedJobs = []seedJob{
	{Job{Title: "Backend Engineer Intern", Company: "Google", Location: "Bengaluru",
		Kind: "internship", Description: "Six month internship for pre-final and final year students. Go or Python.",
		ApplyLink: "https://careers.example/google-intern"}, "arjun@alumni.ac.in"},

	{Job{Title: "Frontend Engineer", Company: "Razorpay", Location: "Bengaluru",
		Kind: "full-time", Description: "React and TypeScript. Open to 2026 and 2027 graduates.",
		ApplyLink: "https://careers.example/razorpay-fe"}, "neha@alumni.ac.in"},

	{Job{Title: "Data Analyst Intern", Company: "Infosys", Location: "Pune",
		Kind: "internship", Description: "SQL and Python. Good entry point into analytics.",
		ApplyLink: "https://careers.example/infosys-data"}, "priya@alumni.ac.in"},

	{Job{Title: "Site Engineer", Company: "L&T Construction", Location: "Jaipur",
		Kind: "full-time", Description: "Civil graduates for highway projects across Rajasthan.",
		ApplyLink: "https://careers.example/lt-site"}, "vikram@alumni.ac.in"},
}

type seedMentorship struct {
	Student string
	Alumni  string
	Message string
	Status  string
}

var seedMentorships = []seedMentorship{
	{"riya@student.ac.in", "arjun@alumni.ac.in",
		"Hello sir, I am targeting backend roles. Could you guide me on what to build?", "accepted"},
	{"meera@student.ac.in", "neha@alumni.ac.in",
		"Hi, I saw your frontend post. Would love some guidance on interviews.", "pending"},
	{"aditya@student.ac.in", "priya@alumni.ac.in",
		"Ma'am, I want to move from electronics to data science. Can you help?", "pending"},
}

func seedIfEmpty() {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("skipping seed:", err)
		return
	}

	ids := map[string]int64{}

	for _, person := range seedPeople {
		result, err := db.Exec(`INSERT INTO users (name, email, password_hash, role, verified, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			person.Name, person.Email, string(hash), person.Role, person.Verified, now())
		if err != nil {
			log.Println("seed user failed:", err)
			continue
		}

		id, _ := result.LastInsertId()
		ids[person.Email] = id

		p := person.Profile
		db.Exec(`INSERT INTO profiles (user_id, college, branch, graduation_year, company,
			job_title, industry, location, skills, interests, bio, linkedin, open_to_mentor, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, p.College, p.Branch, p.GraduationYear, p.Company, p.JobTitle,
			p.Industry, p.Location, p.Skills, p.Interests, p.Bio, p.LinkedIn,
			p.OpenToMentor, now())
	}

	for _, item := range seedEvents {
		e := item.Event
		result, err := db.Exec(`INSERT INTO events
			(title, description, kind, starts_at, location, link, created_by, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Description, e.Kind, e.StartsAt, e.Location, e.Link,
			ids[item.Creator], now())
		if err != nil {
			continue
		}

		eventID, _ := result.LastInsertId()
		for _, email := range item.Attendee {
			db.Exec(`INSERT OR IGNORE INTO event_rsvps (event_id, user_id) VALUES (?, ?)`,
				eventID, ids[email])
		}
	}

	for _, item := range seedPosts {
		p := item.Post
		result, err := db.Exec(`INSERT INTO posts (author_id, title, body, category, flagged, created_at)
			VALUES (?, ?, ?, ?, 0, ?)`,
			ids[item.Author], p.Title, p.Body, p.Category, now())
		if err != nil {
			continue
		}

		postID, _ := result.LastInsertId()
		for _, c := range item.Comments {
			db.Exec(`INSERT INTO comments (post_id, author_id, body, flagged, created_at)
				VALUES (?, ?, ?, 0, ?)`, postID, ids[c.Author], c.Body, now())
		}
	}

	for _, item := range seedJobs {
		j := item.Job
		db.Exec(`INSERT INTO jobs (posted_by, title, company, location, kind, description, apply_link, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			ids[item.Poster], j.Title, j.Company, j.Location, j.Kind,
			j.Description, j.ApplyLink, now())
	}

	for _, m := range seedMentorships {
		db.Exec(`INSERT INTO mentorships (student_id, alumni_id, message, status, created_at)
			VALUES (?, ?, ?, ?, ?)`,
			ids[m.Student], ids[m.Alumni], m.Message, m.Status, now())
	}

	log.Printf("seeded %d demo accounts, password for all: %s", len(ids), seedPassword)
}
