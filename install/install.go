package install

import "fmt"
import "time"

// Run represents a running installation process.
type Run struct {
	done bool

	uiUpdate chan Update
	config   Settings

	steps []step
}

type step interface {
	Exec(chan Update, *Run) error
	Stage() string
	Name() string
}

// Configure prepares an installation.
func Configure(ch chan Update, config Settings) *Run {
	return &Run{
		uiUpdate: ch,
		config:   config,
		steps: []step{
			&PartitionStep{},
			&ConfigureStep{},
			&InstallStep{},
		},
	}
}

// Start commences an installation.
func (r *Run) Start() error {
	go r.install()
	return nil
}

func (r *Run) install() {
	for _, step := range r.steps {
		r.uiUpdate <- Update{Step: step.Stage()}
		r.uiUpdate <- Update{Msg: step.Name() + "\n", Level: MsgCmd}
		if err := step.Exec(r.uiUpdate, r); err != nil {
			r.uiUpdate <- Update{Msg: fmt.Sprintf("\nStep %q failed! %v\n", step.Name(), err), Level: MsgErr}
			return
		}
		r.uiUpdate <- Update{Msg: "\n", Level: MsgCmd}
	}
	r.done = true
}

func (r *Run) Wait() {
	for {
		if r.done {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
}
