package summarize

type ResumeOutput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	BriefResume  string `json:"briefResume"`
	MediumResume string `json:"mediumResume"`
}
