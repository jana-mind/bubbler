package tui

import (
	"charm.land/bubbletea/v2"
	"github.com/jana-mind/bubbler/internal/model"
)

type TuiMsg interface{}

type BoardLoaded struct{ Board model.BoardFile }
type BoardLoadFailed struct{ Err error }

type IssueFocused struct{ IssueID string }
type ColumnFocused struct{ Index int }

type OpenCreateModal struct{}
type OpenMoveModal struct{ IssueID string }
type OpenEditModal struct{ IssueID string }

type FormTitleChanged struct{ Text string }
type FormColumnChanged struct{ Index int }
type FormTagsChanged struct{ Tags []string }
type FormDescLineAdded struct{ Line string }
type FormDescConfirmed struct{}

type ConfirmDelete struct{ IssueID string }
type ConfirmDeleteConfirmed struct{ IssueID string }
type ConfirmDeleteCancelled struct{}

type TagFilterApplied struct{ Tag string }
type TagFilterCleared struct{}

type RefreshRequested struct{}

type WindowResized struct {
	Width  int
	Height int
}

type WriteCompleted struct{ Err error }
type WriteRetryRequested struct{}
type WriteCancelled struct{}
type QuitConfirmed struct{}

var _ tea.Msg = (BoardLoaded{})
var _ tea.Msg = (BoardLoadFailed{})
var _ tea.Msg = (IssueFocused{})
var _ tea.Msg = (ColumnFocused{})
var _ tea.Msg = (OpenCreateModal{})
var _ tea.Msg = (OpenMoveModal{})
var _ tea.Msg = (OpenEditModal{})
var _ tea.Msg = (FormTitleChanged{})
var _ tea.Msg = (FormColumnChanged{})
var _ tea.Msg = (FormTagsChanged{})
var _ tea.Msg = (FormDescLineAdded{})
var _ tea.Msg = (FormDescConfirmed{})
var _ tea.Msg = (ConfirmDelete{})
var _ tea.Msg = (ConfirmDeleteConfirmed{})
var _ tea.Msg = (ConfirmDeleteCancelled{})
var _ tea.Msg = (TagFilterApplied{})
var _ tea.Msg = (TagFilterCleared{})
var _ tea.Msg = (RefreshRequested{})
var _ tea.Msg = (WindowResized{})
var _ tea.Msg = (WriteCompleted{})
var _ tea.Msg = (WriteRetryRequested{})
var _ tea.Msg = (WriteCancelled{})
var _ tea.Msg = (QuitConfirmed{})
