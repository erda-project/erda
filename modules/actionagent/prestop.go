package actionagent

func (agent *Agent) PreStop() {
	// TODO invoke /opt/action/pre-stop

	// agent cancel context to stop other running things
	agent.Cancel()

	agent.stopWatchFiles()

	// 打包目录并上传
	agent.uploadDir()
}
