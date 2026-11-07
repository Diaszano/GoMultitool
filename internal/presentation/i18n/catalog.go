// Package i18n contains the human-facing presentation strings.
package i18n

import (
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// Locale identifies a supported interface language.
type Locale string

const (
	// LocalePtBR is Brazilian Portuguese.
	LocalePtBR Locale = "pt-BR"
	// LocaleEN is English.
	LocaleEN Locale = "en"
)

// MenuItemID identifies a GoMultitool menu entry.
type MenuItemID string

const (
	// MenuHardware identifies the hardware information entry.
	MenuHardware MenuItemID = "hardware"
	// MenuStorage identifies the future storage entry.
	MenuStorage MenuItemID = "storage"
	// MenuBackup identifies the future backup entry.
	MenuBackup MenuItemID = "backup"
	// MenuRestore identifies the future restore entry.
	MenuRestore MenuItemID = "restore"
	// MenuInstall identifies the future install entry.
	MenuInstall MenuItemID = "install"
	// MenuConfigureImage identifies the future image configuration entry.
	MenuConfigureImage MenuItemID = "configure-image"
	// MenuDiagnostics identifies the future diagnostics entry.
	MenuDiagnostics MenuItemID = "diagnostics"
	// MenuSystemSettings identifies the future system settings entry.
	MenuSystemSettings MenuItemID = "system-settings"
	// MenuTerminal identifies the future terminal entry.
	MenuTerminal MenuItemID = "terminal"
	// MenuExit identifies the exit entry.
	MenuExit MenuItemID = "exit"
)

// MenuItem is a localized menu label and its secondary hint.
type MenuItem struct {
	ID    MenuItemID
	Label string
	Hint  string
}

// Catalog groups all application-authored text shown to a person.
type Catalog struct {
	ProductName      string
	Menu             [10]MenuItem
	Navigation       string
	HardwareTitle    string
	EmptyInventory   string
	ErrorPrefix      string
	InUseLabel       string
	BootMediaLabel   string
	DecimalSeparator string
	Kinds            map[device.Kind]string
	ConsolePrompt    string
	InvalidOption    string
	CommitLabel      string
	BuildLabel       string
	Usage            string
}

// DetectLocale selects the first supported locale from the standard environment order.
func DetectLocale(getenv func(string) string) Locale {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		value := strings.TrimSpace(getenv(key))
		value = strings.SplitN(value, "@", 2)[0]
		value = strings.SplitN(value, ".", 2)[0]
		language := strings.ToLower(strings.SplitN(strings.ReplaceAll(value, "-", "_"), "_", 2)[0])
		switch language {
		case "pt":
			return LocalePtBR
		case "en":
			return LocaleEN
		}
	}
	return LocalePtBR
}

// For returns the catalog for locale. Unknown values use Brazilian Portuguese.
func For(locale Locale) Catalog {
	if locale == LocaleEN {
		return Catalog{
			ProductName: "GoMultitool",
			Menu: [10]MenuItem{
				{MenuHardware, "Hardware information", "Inspect connected storage devices"},
				{MenuStorage, "Storage", "Coming soon"},
				{MenuBackup, "Backup", "Coming soon"},
				{MenuRestore, "Restore", "Coming soon"},
				{MenuInstall, "Install Linux", "Coming soon"},
				{MenuConfigureImage, "Configure image", "Coming soon"},
				{MenuDiagnostics, "Diagnostics", "Coming soon"},
				{MenuSystemSettings, "System settings", "Coming soon"},
				{MenuTerminal, "Terminal", "Coming soon"},
				{MenuExit, "Exit", "Leave GoMultitool"},
			},
			Navigation:       "↑/↓ Move • Enter Select • Esc Back • q Quit",
			HardwareTitle:    "Storage devices",
			EmptyInventory:   "No devices found.",
			ErrorPrefix:      "Error: ",
			InUseLabel:       "in use",
			BootMediaLabel:   "boot media",
			DecimalSeparator: ".",
			Kinds:            kinds("unknown"),
			ConsolePrompt:    "Choose: ",
			InvalidOption:    "Invalid option.",
			CommitLabel:      "Commit",
			BuildLabel:       "Build",
			Usage:            "Usage: multitool version | multitool hardware --json | multitool diagnose --json | multitool tui | multitool --console",
		}
	}
	return Catalog{
		ProductName: "GoMultitool",
		Menu: [10]MenuItem{
			{
				MenuHardware,
				"Informações de hardware",
				"Ver dispositivos de armazenamento conectados",
			},
			{MenuStorage, "Armazenamento", "Em breve"},
			{MenuBackup, "Backup", "Em breve"},
			{MenuRestore, "Restaurar", "Em breve"},
			{MenuInstall, "Instalar Linux", "Em breve"},
			{MenuConfigureImage, "Configurar imagem", "Em breve"},
			{MenuDiagnostics, "Diagnósticos", "Em breve"},
			{MenuSystemSettings, "Configurações do sistema", "Em breve"},
			{MenuTerminal, "Terminal", "Em breve"},
			{MenuExit, "Sair", "Fechar o GoMultitool"},
		},
		Navigation:       "↑/↓ navegar • Enter selecionar • Esc voltar • q sair",
		HardwareTitle:    "Dispositivos de armazenamento",
		EmptyInventory:   "Nenhum dispositivo encontrado.",
		ErrorPrefix:      "Erro: ",
		InUseLabel:       "em uso",
		BootMediaLabel:   "mídia de boot",
		DecimalSeparator: ",",
		Kinds:            kinds("desconhecido"),
		ConsolePrompt:    "Escolha: ",
		InvalidOption:    "Opção inválida.",
		CommitLabel:      "Commit",
		BuildLabel:       "Compilação",
		Usage:            "Uso: multitool version | multitool hardware --json | multitool diagnose --json | multitool tui | multitool --console",
	}
}

func kinds(unknown string) map[device.Kind]string {
	return map[device.Kind]string{
		device.KindEMMC:    "eMMC",
		device.KindSD:      "SD",
		device.KindUSB:     "USB",
		device.KindNAND:    "NAND",
		device.KindUnknown: unknown,
	}
}
