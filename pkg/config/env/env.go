package pkgenv

import (
	"os"
	"strings"
)

// Environment representa el entorno funcional de la app (dev, stg, prod)
type Environment int

// Platform representa la plataforma donde corre la app (local, docker, mix, fury, aws, gcp)
type Platform int

// Enumeración de ambientes
const (
	Dev Environment = iota
	Stg
	Prod
)

// Enumeración de plataformas
const (
	Local Platform = iota
	Mix            // API local, dependencias en Docker
	Docker
	Fury
	AWS
	GCP
)

// Nombres legibles para los enums
var envNames = [...]string{
	"dev",
	"stg",
	"prod",
}

var platformNames = [...]string{
	"local",
	"docker",
	"mix",
	"fury",
	"aws",
	"gcp",
}

func (e Environment) String() string {
	if e < Dev || e > Prod {
		return envNames[Dev]
	}
	return envNames[e]
}

func (p Platform) String() string {
	if p < Local || p > GCP {
		return platformNames[Local]
	}
	return platformNames[p]
}

func GetEnvFromString(s string) Environment {
	switch strings.ToLower(s) {
	case "prod":
		return Prod
	case "stg":
		return Stg
	case "dev":
		return Dev
	default:
		return Dev // Fallback
	}
}

func GetPlatformFromString(s string) Platform {
	switch strings.ToLower(s) {
	case "docker":
		return Docker
	case "mix":
		return Mix
	case "fury":
		return Fury
	case "aws":
		return AWS
	case "gcp":
		return GCP
	case "local":
		fallthrough
	default:
		return Local
	}
}

func FromEnv() (env Environment, platform Platform) {
	env = GetEnvFromString(os.Getenv("DEPLOY_ENV"))
	platform = GetPlatformFromString(os.Getenv("DEPLOY_PLATFORM"))
	return
}

func IsProd(env Environment) bool     { return env == Prod }
func IsStg(env Environment) bool      { return env == Stg }
func IsDev(env Environment) bool      { return env == Dev }
func IsMix(platform Platform) bool    { return platform == Mix }
func IsLocal(platform Platform) bool  { return platform == Local }
func IsDocker(platform Platform) bool { return platform == Docker }
func IsFury(platform Platform) bool   { return platform == Fury }

// IsDangerousCombo valida combinaciones peligrosas
func IsDangerousCombo(env Environment, platform Platform) bool {
	// Por ejemplo: nunca deberías correr "prod" en mix/local
	if env == Prod && (platform == Local || platform == Docker || platform == Mix) {
		return true
	}
	return false
}
