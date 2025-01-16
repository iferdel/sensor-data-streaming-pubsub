package sensorlogic

const (
	CommandLogin                 = "login"
	CommandLogout                = "logout"
	CommandSleep                 = "sleep"
	CommandAwake                 = "awake"
	CommandChangeSampleFrequency = "changeSampleFrequency"
	CommandDelete                = "delete"
)

var ValidCommands = map[string]bool{
	CommandLogin:                 true,
	CommandLogout:                true,
	CommandSleep:                 true,
	CommandAwake:                 true,
	CommandChangeSampleFrequency: true,
	CommandDelete:                true,
}
