package components

import "github.com/jana-mind/bubbler/internal/model"

type MainViewModel struct {
	ViewState     int
	Board         model.BoardFile
	Issues        map[string]model.IssueFile
	FocusedColumn int
	FocusedIssue  int
	DetailIssueID string
	ConfirmDelete bool
	FormTitle     string
	FormColumn    int
	FormTags      []string
	FormDescLines []string
	TagFilter     string
	WriteErr      error
	Loading       bool
	BoardName     string
	Width         int
	Height        int
	Completion   CompletionViewModel
}

type CompletionViewModel struct {
	Active  bool
	Matches []string
	Index   int
}

func RenderView(m MainViewModel) string {
	if m.Loading {
		return "Loading...\n"
	}
	if m.BoardName == "" {
		return "No board found. Run `bubbler init` first.\nPress any key to exit.\n"
	}
	switch m.ViewState {
	case 0:
		return RenderBoard(TUIViewModel{
			Board:         m.Board,
			FocusedColumn: m.FocusedColumn,
			FocusedIssue:  m.FocusedIssue,
		}, m.Width) + "\n" + RenderStatusbar(m.BoardName, len(m.Board.Issues), m.TagFilter)
	case 1:
		if m.ConfirmDelete {
			return RenderDetail(TUIDetailModel{
				Board:         m.Board,
				Issues:        m.Issues,
				DetailIssueID: m.DetailIssueID,
			}) + "\n" + RenderConfirm(m.DetailIssueID)
		}
		return RenderDetail(TUIDetailModel{
			Board:         m.Board,
			Issues:        m.Issues,
			DetailIssueID: m.DetailIssueID,
		})
	case 2:
		return RenderCreate(TUICreateModel{
			Board:      m.Board,
			FormTitle:  m.FormTitle,
			FormColumn: m.FormColumn,
			FormTags:   m.FormTags,
			Completion: m.Completion,
		})
	case 3:
		return RenderMove(TUIMoveModel{
			Board:         m.Board,
			FormColumn:    m.FormColumn,
			DetailIssueID: m.DetailIssueID,
		})
	case 4:
		return RenderEdit(TUIEditModel{
			Board:         m.Board,
			Issues:        m.Issues,
			DetailIssueID: m.DetailIssueID,
			FormTitle:     m.FormTitle,
			FormDescLines: m.FormDescLines,
			FormTags:      m.FormTags,
			Completion:   m.Completion,
		})
	case 5:
		return RenderFilter(TUIFilterModel{
			Board:     m.Board,
			TagFilter: m.TagFilter,
		}, m.Completion)
	default:
		return "Unknown view\n"
	}
}
