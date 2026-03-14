package config

// Core configuration file and directory names.
const (
	ConfigDirName  = ".echarge-report"
	ConfigFileName = ".echargereport.yaml"
)

// Configuration keys – central constants for all Viper lookups.
// To rename a key, a single change here is sufficient.
const (
	// General wallbox configuration
	KeyWallbox     = "wallbox"

	// Wallbox connection configuration (Nested)
	KeyWallboxChipIds          = "chipIds"
	KeyWallboxGoeCloudToken    = "wallbox.goe.cloud.token"
	KeyWallboxGoeCloudSerial   = "wallbox.goe.cloud.serial"
	KeyWallboxGoeLocalApiUrl   = "wallbox.goe.local.apiUrl"

	// General report configuration
	KeyLicensePlate = "licenseplate"
	KeyKwhPrice     = "kwhprice"

	// Home Assistant configuration
	KeyHAToken        = "smarthome.homeassistant.token"
	KeyHAAPI          = "smarthome.homeassistant.api"
	KeyHAMilageSensor = "smarthome.homeassistant.milage_sensorid"

	// Mail configuration
	KeyMailHost     = "mail.host"
	KeyMailPort     = "mail.port"
	KeyMailUsername = "mail.username"
	KeyMailPassword = "mail.password"
	KeyMailFrom     = "mail.from"
	KeyMailTo       = "mail.to"
)
