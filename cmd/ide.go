package cmd

func runIDE(sessionName string) error {
	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	tmux("new-session", "-d", "-s", sessionName)
	tmux("split-window", "-h", "-t", sessionName+":0")
	tmux("split-window", "-v", "-t", sessionName+":0.1")
	tmux("send-keys", "-t", sessionName+":0.0", "vim", "Enter")
	tmux("select-pane", "-t", sessionName+":0.0")

	return attachSession(sessionName)
}
