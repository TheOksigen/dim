package results

type Subject struct {
	Name    string `json:"name"`
	Score   int16  `json:"score"`
	Answers string `json:"answers"`
}

type Result struct {
	FINCode    string    `json:"finCode"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	FatherName string    `json:"fatherName"`
	ExamDate   string    `json:"examDate"`
	Subjects   []Subject `json:"subjects"`
	TotalScore int16     `json:"totalScore"`
	Passed     bool      `json:"passed"`
}
