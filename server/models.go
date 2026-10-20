package main

// models.go — Go structs that mirror the database tables.
//
// The `json:"..."` tags decide the key names the Electron app sees. Keeping
// them snake_case means the JSON looks exactly like the SQL columns, so there
// is only one set of names to remember.

// User is a row from the users table. PasswordHash has json:"-" so it is never
// accidentally sent to the client.
type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Role      string `json:"role"`
	Verified  bool   `json:"verified"`
	CreatedAt string `json:"created_at"`
}

// Profile is a row from the profiles table.
type Profile struct {
	UserID         int64  `json:"user_id"`
	College        string `json:"college"`
	Branch         string `json:"branch"`
	GraduationYear int    `json:"graduation_year"`
	Company        string `json:"company"`
	JobTitle       string `json:"job_title"`
	Industry       string `json:"industry"`
	Location       string `json:"location"`
	Skills         string `json:"skills"`
	Interests      string `json:"interests"`
	Bio            string `json:"bio"`
	LinkedIn       string `json:"linkedin"`
	OpenToMentor   bool   `json:"open_to_mentor"`
	UpdatedAt      string `json:"updated_at"`
}

// DirectoryEntry is one person as shown in the alumni directory: their user
// row and their profile row joined together.
type DirectoryEntry struct {
	User
	Profile Profile `json:"profile"`
}

// Recommendation is a suggested mentor plus the score that put them there.
// Reasons is a human readable explanation like ["2 shared skills: Go, React"]
// so the UI can show *why* a match was suggested instead of a magic number.
type Recommendation struct {
	DirectoryEntry
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

// Event is an alumni meetup, webinar or panel discussion.
type Event struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	StartsAt    string `json:"starts_at"`
	Location    string `json:"location"`
	Link        string `json:"link"`
	CreatedBy   int64  `json:"created_by"`
	CreatorName string `json:"creator_name"`
	CreatedAt   string `json:"created_at"`
	RSVPCount   int    `json:"rsvp_count"`
	IsAttending bool   `json:"is_attending"`
}

// Post is a discussion forum thread.
type Post struct {
	ID           int64  `json:"id"`
	AuthorID     int64  `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorRole   string `json:"author_role"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Category     string `json:"category"`
	Flagged      bool   `json:"flagged"`
	CreatedAt    string `json:"created_at"`
	CommentCount int    `json:"comment_count"`
}

// Comment is a reply on a forum post.
type Comment struct {
	ID         int64  `json:"id"`
	PostID     int64  `json:"post_id"`
	AuthorID   int64  `json:"author_id"`
	AuthorName string `json:"author_name"`
	AuthorRole string `json:"author_role"`
	Body       string `json:"body"`
	Flagged    bool   `json:"flagged"`
	CreatedAt  string `json:"created_at"`
}

// Mentorship is a student's request for an alumnus to mentor them.
type Mentorship struct {
	ID          int64  `json:"id"`
	StudentID   int64  `json:"student_id"`
	StudentName string `json:"student_name"`
	AlumniID    int64  `json:"alumni_id"`
	AlumniName  string `json:"alumni_name"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// Job is a vacancy posted by an alumnus for placement assistance.
type Job struct {
	ID          int64  `json:"id"`
	PostedBy    int64  `json:"posted_by"`
	PosterName  string `json:"poster_name"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	ApplyLink   string `json:"apply_link"`
	CreatedAt   string `json:"created_at"`
}
