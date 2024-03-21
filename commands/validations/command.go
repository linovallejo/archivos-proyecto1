package Command

import (
	"errors"
	"strings"
)

// CommandParamDetails holds information about whether a parameter is mandatory or optional
type CommandParamDetails struct {
	IsMandatory bool
}

// commandParams maps each command to a map of its parameters and their details
var commandParams = map[string]map[string]CommandParamDetails{
	"mkdisk": {
		"size": {IsMandatory: true},
		"unit": {IsMandatory: false},
		"fit":  {IsMandatory: false},
	},
	"rmdisk": {
		"driveletter": {IsMandatory: true},
	},
	"fdisk": {
		"size":        {IsMandatory: false},
		"driveletter": {IsMandatory: true},
		"name":        {IsMandatory: true},
		"unit":        {IsMandatory: false},
		"type":        {IsMandatory: false},
		"fit":         {IsMandatory: false},
		"delete":      {IsMandatory: false},
		"add":         {IsMandatory: false},
	},
	"mount": {
		"driveletter": {IsMandatory: true},
		"name":        {IsMandatory: true},
	},
	"unmount": {
		"id": {IsMandatory: true},
	},
	"mkfs": {
		"id":   {IsMandatory: true},
		"type": {IsMandatory: false},
		"fs":   {IsMandatory: false},
	},
	"rep": {
		"name": {IsMandatory: true},
		"path": {IsMandatory: true},
		"id":   {IsMandatory: true},
		"ruta": {IsMandatory: false},
	},
}

func ValidarParametros(commandLine string) error {
	hashIndex := strings.Index(commandLine, "#")
	if hashIndex != -1 {
		commandLine = commandLine[:hashIndex]
	}
	parts := strings.Fields(commandLine)
	if len(parts) < 2 {
		return errors.New("formato de comando invalido")
	}
	cmd := parts[0]
	allowedParams, ok := commandParams[cmd]
	if !ok {
		return errors.New("comando no especificado: " + cmd)
	}

	// Track provided parameters to check mandatory ones at the end
	providedParams := make(map[string]bool)

	for _, part := range parts[1:] {
		p := strings.Split(part, "=")
		if len(p) != 2 {
			return errors.New("formato de parametro invalido: " + part)
		}
		paramName := p[0][1:] // Remove "-" from param name

		// Check if the parameter is allowed for this command
		if _, ok := allowedParams[paramName]; !ok {
			return errors.New("parametro no especificado para " + cmd + ": " + paramName)
		}

		providedParams[paramName] = true
	}

	// Check for missing mandatory parameters
	for paramName, details := range allowedParams {
		if details.IsMandatory && !providedParams[paramName] {
			return errors.New("parametro obligatorio faltante para " + cmd + ": " + paramName)
		}
	}

	return nil
}

var CommandsNotImplemented = []string{
	"login",
	"logout",
	"mkgrp",
	"rmgrp",
	"mkusr",
	"rmusr",
	"mkfile",
	"cat",
	"remove",
	"edit",
	"rename",
	"mkdir",
	"copy",
	"move",
	"find",
	"chown",
	"chgrp",
	"chmod",
	"loss",
	"recovery",
}

func ValidarComando(commandLine string) error {
	parts := strings.Fields(commandLine)
	if len(parts) < 1 {
		return errors.New("formato de comando invalido")
	}
	var cmd string = strings.TrimSpace(parts[0])
	cmd = strings.ToLower(cmd)
	if strings.Contains(strings.Join(CommandsNotImplemented, ","), cmd) {
		return errors.New("comando no implementado: " + cmd)
	}
	return nil
}
