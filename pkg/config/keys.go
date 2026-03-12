package config

// Core configuration file and directory names.
const (
	ConfigDirName  = ".echarge-report"
	ConfigFileName = ".echargereportrc"
)

// Configuration keys – central constants for all Viper lookups.
// To rename a key, a single change here is sufficient.
const (
	// General wallbox configuration
	KeyWallboxType = "wallbox_type" // Type of wallbox: "goe", "easee", etc.

	// Wallbox connection configuration
	KeyWallboxToken       = "wallbox_token"
	KeyWallboxLocalApiUrl = "wallbox_localApiUrl"
	KeyWallboxSerial      = "wallbox_serial"
	KeyWallboxChipIds     = "wallbox_chipIds"

	// General report configuration
	KeyLicensePlate = "licenseplate"
	KeyKwhPrice     = "kwhprice"

	// Home Assistant configuration
	KeyHAToken        = "ha_token"
	KeyHAAPI          = "ha_api"
	KeyHAMilageSensor = "ha_milage_sensorid"

	// Mail configuration
	KeyMailHost     = "mail_host"
	KeyMailPort     = "mail_port"
	KeyMailUsername = "mail_username"
	KeyMailPassword = "mail_password"
	KeyMailFrom     = "mail_from"
	KeyMailTo       = "mail_to"
)
