package main

type settings struct {
	Username string
	Hostname string
	Password string
	Timezone string

	Disk  disk
	Scrub bool
}
