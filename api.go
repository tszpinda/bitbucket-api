package bapi

type PullRequestStatus int

func (status PullRequestStatus) String() string {
	switch status {
	case PullRequestOpen:
		return "OPEN"
	case PullRequestDeclined:	
		return "DECLINED"
	case PullRequestMerged:
		return "MERGED"
	}
	panic("invalid status")
}

const (
	PullRequestOpen PullRequestStatus = iota
	PullRequestMerged
	PullRequestDeclined
)

type PullRequestList struct {
	Pagelen      int   `json:"pagelen"`
	Page         int   `json:"page"`
	Size         int   `json:"size"`
	Links        Links `json:"links"`
	PullRequests []struct {
		Id          int      `json:"id"`
		Description string      `json:"description"`
		Title       string      `json:"title"`
		State       string      `json:"state"`
		CreatedOn   string      `json:"created_on"`
		UpdatedOn   string      `json:"updated_on"`
		Type        string      `json:"type"`
		Author      Author      `json:"author"`
		Destination Destination `json:"destination"`
		Source      Source      `json:"source"`
	} `json:"values"`
}

type PullRequest struct {
	Description       string   `json:"description"`
	Title             string   `json:"title"`
	Links             Links    `json:"links"`
	CloseSourceBranch bool     `json:"close_source_branch"`
	Reviewers         Reviewer `json:"reviewers"`
	MergeCommit  string        `json:"merge_commit"`
	Reason       string        `json:"reason"`
	ClosedBy     string        `json:"closed_by"`
	Source       Source        `json:"source"`
	State        string        `json:"state"`
	Author       Author        `json:"author"`
	CreatedOn    string        `json:"created_on"`
	UpdatedOn    string        `json:"updated_on"`
	Type         string        `json:"type"`
	Id           int        `json:"id"`
	Destination  Destination   `json:"destination"`
	Participants []Participant `json:"participants"`
}

type Participant struct {
	Role     string `json:"role"`
	User     Author `json:"user"`
	Approved bool   `json:"approved"`
}

type Author struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	UUID        string `json:"uuid"`
}

type Links struct {
	Decline  Link `json:"decline"`
	Commits  Link `json:"commits"`
	Self     Link `json:"self"`
	Comments Link `json:"comments"`
	Merge    Link `json:"merge"`
	Html     Link `json:"html"`
	Activity Link `json:"activity"`
	Diff     Link `json:"diff"`
	Approve  Link `json:"approve"`
}

type Link struct {
	Href string `json:"href"`
}

type Reviewer struct {
}

type Destination struct {
	Branch struct {
		Name string `json:"name"`
	} `json:"branch"`
}
type Source struct {
	Branch struct {
		Name string `json:"name"`
	} `json:"branch"`
}
