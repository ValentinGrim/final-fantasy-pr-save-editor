package character

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Commands struct {
		widget.BaseWidget
		enabled binding.Bool
		inputs  []*inputs.IdEntry
		job     *inputs.IdEntry
	}
)

func NewCoreCommands(c *core.Character) *Commands {
	e := &Commands{
		enabled: binding.BindBool(&c.EnableCommandsSave),
		inputs:  make([]*inputs.IdEntry, len(c.Commands)),
		job:     inputs.NewIdEntryWithDataWithHint(&c.Base.JobID, finder.Jobs),
	}
	e.ExtendBaseWidget(e)

	for i := 0; i < len(c.Commands); i++ {
		j := inputs.NewIdEntryWithDataWithHint(&c.Commands[i], finder.Commands)
		j.ID.Disable()
		e.inputs[i] = j
	}
	e.enabled.AddListener(e)
	return e
}

func (e *Commands) DataChanged() {
	enabled, _ := e.enabled.Get()
	for _, i := range e.inputs {
		if enabled {
			i.ID.Enable()
		} else {
			i.ID.Disable()
		}
	}
	if enabled {
		e.job.ID.Enable()
	} else {
		e.job.ID.Disable()
	}
}

// CreateRenderer shows a form of the job and the commands, the searches on the
// right.
func (e *Commands) CreateRenderer() fyne.WidgetRenderer {
	rows := inputs.NewForm(widget.NewLabel("Job:"), inputs.IDValue(e.job.ID, e.job.Label))
	for i, in := range e.inputs {
		label := ""
		if i == 0 {
			label = "Commands:"
		}
		rows.Add(widget.NewLabel(label))
		rows.Add(inputs.IDValue(in.ID, in.Label))
	}
	top := container.NewVBox(
		widget.NewLabel("can cause soft locks when loading save"),
		widget.NewCheckWithData("Enabled", e.enabled))
	return widget.NewSimpleRenderer(container.NewBorder(top, nil, nil, nil,
		inputs.WithSearch(container.NewVScroll(rows), inputs.SearchPanels(
			inputs.NamedSearch{Name: "Commands", Search: inputs.GetSearches().Commands},
			inputs.NamedSearch{Name: "Jobs", Search: inputs.GetSearches().Jobs}))))
}
