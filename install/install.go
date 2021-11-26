package install

// Run represents a running installation process.
type Run struct {
	uiUpdate chan Update
	config   Settings
}

// Configure prepares an installation.
func Configure(ch chan Update, config Settings) *Run {
	return &Run{
		uiUpdate: ch,
		config:   config,
	}
}

// Start commences an installation.
func (r *Run) Start() error {
	r.uiUpdate <- Update{Step: "format"}
	go r.install()
	return nil
}

func (r *Run) install() {
	r.uiUpdate <- Update{Msg: "Commencing installation.\n", Level: MsgCmd}

	runCmdInteractive(r.uiUpdate, "sys", "uname", "-a")
}
