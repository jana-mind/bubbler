package components

import (
	"github.com/jana-mind/bubbler/internal/model"
)

type TUIViewModel struct {
	Board         model.BoardFile
	FocusedColumn int
	FocusedIssue  int
}

type TUIDetailModel struct {
	Board         model.BoardFile
	Issues        map[string]model.IssueFile
	DetailIssueID string
}

type TUICreateModel struct {
	Board      model.BoardFile
	FormTitle  string
	FormColumn int
	FormTags   []string
	TagInput   string
	Completion CompletionViewModel
}

type TUIMoveModel struct {
	Board         model.BoardFile
	FormColumn    int
	DetailIssueID string
}

type TUIEditModel struct {
	Board         model.BoardFile
	Issues        map[string]model.IssueFile
	DetailIssueID string
	FormTitle     string
	FormDescLines []string
	FormTags      []string
	TagInput      string
	Completion    CompletionViewModel
}

type TUIFilterModel struct {
	Board     model.BoardFile
	TagFilter string
}

type TUIConfirmModel struct {
	DetailIssueID string
}
