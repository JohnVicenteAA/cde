package cmd

func runIDE(sessionName string) error {
	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	runner.Run("new-session", "-d", "-s", sessionName)
	runner.Run("split-window", "-h", "-t", sessionName+":0")
	runner.Run("split-window", "-v", "-t", sessionName+":0.1")
	runner.Run("send-keys", "-t", sessionName+":0.0", "nvim", "Enter")
	runner.Run("select-pane", "-t", sessionName+":0.0")

	return runner.Attach(sessionName)
}
