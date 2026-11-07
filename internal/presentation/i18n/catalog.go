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
	// MenuBurn identifies the image burn entry.
	MenuBurn MenuItemID = "burn"
	// MenuBackup identifies the backup entry.
	MenuBackup MenuItemID = "backup"
	// MenuHardware identifies the hardware information entry.
	MenuHardware MenuItemID = "hardware"
	// MenuLegacy identifies the legacy Bash menu entry.
	MenuLegacy MenuItemID = "legacy"
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
	Menu             [5]MenuItem
	Navigation       string
	HardwareTitle    string
	EmptyInventory   string
	ErrorPrefix      string
	Loading          string
	Busy             string
	OK               string
	InUseLabel       string
	BootMediaLabel   string
	DecimalSeparator string
	Kinds            map[device.Kind]string
	ConsolePrompt    string
	InvalidOption    string
	CommitLabel      string
	BuildLabel       string
	Usage            string
	Burn             BurnTexts
}

// BurnTexts are the strings of the image burn screen.
type BurnTexts struct {
	Destination, Images, SizeAtBurn, Warning, Start, Cancel string
	Success, CommandRate, NoImages, NoDestination           string
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
			Menu: [5]MenuItem{
				{MenuBurn, "Burn image to eMMC", "Install Armbian or another image"},
				{MenuBackup, "Back up", "Save the eMMC to the MULTITOOL partition"},
				{MenuHardware, "Hardware information", "Inspect connected storage devices"},
				{MenuLegacy, "Legacy menu (Bash)", "Restore, erase, steP-nand"},
				{MenuExit, "Exit", "Leave GoMultitool"},
			},
			Navigation:       "↑/↓ Move • Enter Select • Esc Back • q Quit",
			HardwareTitle:    "Storage devices",
			EmptyInventory:   "No devices found.",
			ErrorPrefix:      "Error: ",
			Loading:          "Loading…",
			Busy:             "Another operation is running.",
			OK:               "OK",
			InUseLabel:       "in use",
			BootMediaLabel:   "boot media",
			DecimalSeparator: ".",
			Kinds:            kinds("unknown"),
			ConsolePrompt:    "Choose: ",
			InvalidOption:    "Invalid option.",
			CommitLabel:      "Commit",
			BuildLabel:       "Build",
			Usage:            "Usage: multitool version | multitool hardware --json | multitool diagnose --json | multitool tui | multitool --console",
			Burn: BurnTexts{
				Destination:   "Destination",
				Images:        "Image",
				SizeAtBurn:    "size computed when burning",
				Warning:       "Everything on %s will be erased.",
				Start:         "Burn",
				Cancel:        "Cancel",
				Success:       "Burned and verified. You can power off.",
				CommandRate:   "The image has no DDR command rate; the previous value was not applied.",
				NoImages:      "No images in MULTITOOL/images.",
				NoDestination: "No available destination.",
			},
		}
	}
	return Catalog{
		ProductName: "GoMultitool",
		Menu: [5]MenuItem{
			{MenuBurn, "Gravar imagem no eMMC", "Instalar Armbian ou outra imagem"},
			{MenuBackup, "Fazer backup", "Salvar o eMMC na partição MULTITOOL"},
			{
				MenuHardware,
				"Informações de hardware",
				"Ver dispositivos de armazenamento conectados",
			},
			{MenuLegacy, "Menu antigo (Bash)", "Restaurar, apagar, steP-nand"},
			{MenuExit, "Sair", "Fechar o GoMultitool"},
		},
		Navigation:       "↑/↓ navegar • Enter selecionar • Esc voltar • q sair",
		HardwareTitle:    "Dispositivos de armazenamento",
		EmptyInventory:   "Nenhum dispositivo encontrado.",
		ErrorPrefix:      "Erro: ",
		Loading:          "Carregando…",
		Busy:             "Outra operação em andamento.",
		OK:               "OK",
		InUseLabel:       "em uso",
		BootMediaLabel:   "mídia de boot",
		DecimalSeparator: ",",
		Kinds:            kinds("desconhecido"),
		ConsolePrompt:    "Escolha: ",
		InvalidOption:    "Opção inválida.",
		CommitLabel:      "Commit",
		BuildLabel:       "Compilação",
		Usage:            "Uso: multitool version | multitool hardware --json | multitool diagnose --json | multitool tui | multitool --console",
		Burn: BurnTexts{
			Destination:   "Destino",
			Images:        "Imagem",
			SizeAtBurn:    "tamanho calculado ao gravar",
			Warning:       "Tudo em %s será apagado.",
			Start:         "Gravar",
			Cancel:        "Cancelar",
			Success:       "Gravado e conferido. Pode desligar.",
			CommandRate:   "A imagem não tem o command rate da DDR; o valor anterior não foi aplicado.",
			NoImages:      "Nenhuma imagem em MULTITOOL/images.",
			NoDestination: "Nenhum destino disponível.",
		},
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
