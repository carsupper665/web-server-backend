package services

import "fmt"

// These service functions are placeholders. Replace them with real logic.

func StartServer() {
	fmt.Println("start server")
}

func StopServer() {
	fmt.Println("stop server")
}

func SwitchVersion(version string) {
	fmt.Println("switch version to", version)
}

func BackupWorld() {
	fmt.Println("backup world")
}

func ReadLogs() string {
	// TODO: read actual log file
	return "logs..."
}

func AddMod(name string) {
	fmt.Println("add mod", name)
}

func DeleteMod(name string) {
	fmt.Println("delete mod", name)
}
