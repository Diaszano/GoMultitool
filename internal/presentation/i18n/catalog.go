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
	// MenuReboot identifies the restart action.
	MenuReboot MenuItemID = "reboot"
	// MenuPowerOff identifies the shutdown action.
	MenuPowerOff MenuItemID = "poweroff"
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
	Menu             [7]MenuItem
	HomeTitle        string
	HomeNav          string
	FormNav          string
	HardwareNav      string
	HardwareTitle    string
	EmptyInventory   string
	ErrorPrefix      string
	Loading          string
	Busy             string
	ImageTooLarge    string
	DeviceChanged    string
	UnsafeDevice     string
	ReadbackMismatch string
	NotSelectable    string
	Unavailable      string
	OK               string
	ConfirmReboot    string
	ConfirmPowerOff  string
	PowerUnavailable string
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
	Backup           BackupTexts
}

// BackupTexts are the strings of the backup screen.
type BackupTexts struct {
	Title, PickSource                             string
	Source, Name, Free, Start, Cancel, Cancelled  string
	Saved, InvalidName, NoSpace, Exists, NoSource string
}

// BurnTexts are the strings of the image burn screen.
type BurnTexts struct {
	Title, Review, PickBoth, PickDevice, PickImage, Size     string
	Destination, Images, SizeAtBurn, Warning, Start, Cancel  string
	Success, CommandRate, NoImages, NoDestination, Cancelled string
	Sizing, Verifying                                        string
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
			Menu: [7]MenuItem{
				{MenuBurn, "Burn image to eMMC", "Install Armbian or another image"},
				{MenuBackup, "Back up", "Save the eMMC to the MULTITOOL partition"},
				{MenuHardware, "Hardware information", "Inspect connected storage devices"},
				{MenuLegacy, "Legacy menu (Bash)", "Restore, erase, steP-nand"},
				{MenuReboot, "Reboot", "Restart the device"},
				{MenuPowerOff, "Power off", "Shut down the device"},
				{MenuExit, "Exit", "Leave GoMultitool"},
			},
			HomeTitle:        "Recovery menu",
			HomeNav:          "↑/↓ or ←/→ Move  ·  Enter Open  ·  q Quit",
			FormNav:          "Tab/Shift+Tab or ←/→ Focus  ·  ↑/↓ Move  ·  Enter Select  ·  Esc Back",
			HardwareNav:      "↑/↓ Scroll  ·  Esc Back  ·  q Quit",
			HardwareTitle:    "Storage devices",
			EmptyInventory:   "No devices found.",
			ErrorPrefix:      "Error: ",
			Loading:          "Loading…",
			Busy:             "Another operation is running.",
			ImageTooLarge:    "The image is larger than the destination.",
			DeviceChanged:    "The device changed after it was chosen; choose it again.",
			UnsafeDevice:     "This device cannot be written to.",
			ReadbackMismatch: "Verification failed: the written data differs from the image. Burn again.",
			NotSelectable:    "not selectable",
			Unavailable:      "unavailable",
			OK:               "OK",
			ConfirmReboot:    "Restart the device now?",
			ConfirmPowerOff:  "Shut down the device now?",
			PowerUnavailable: "Power controls are available when started from the GoMultitool recovery image.",
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
				Title:         "Burn image",
				Review:        "Confirm burn",
				PickBoth:      "Press Enter to select the destination and the image.",
				PickDevice:    "Press Enter on the device you want to erase.",
				PickImage:     "Press Enter to select the image.",
				Size:          "Size",
				Destination:   "Destination",
				Images:        "Image",
				SizeAtBurn:    "size computed when burning",
				Warning:       "Everything on %s will be erased.",
				Start:         "Burn",
				Cancel:        "Cancel",
				Cancelled:     "Burn cancelled. The device is incomplete; burn again before using it.",
				Success:       "Burned and verified. You can power off.",
				CommandRate:   "The image has no DDR command rate; the previous value was not applied.",
				NoImages:      "No images in MULTITOOL/images.",
				NoDestination: "No available destination.",
				Sizing:        "Computing image size…",
				Verifying:     "Verifying the written data…",
			},
			Backup: BackupTexts{
				Title:       "Back up device",
				PickSource:  "Press Enter to select the device to back up.",
				Source:      "Source",
				Name:        "Name",
				Free:        "Free on MULTITOOL: %s",
				Start:       "Back up",
				Cancel:      "Cancel",
				Cancelled:   "Backup cancelled. No file was saved.",
				Saved:       "Backup saved: %s.img.gz",
				InvalidName: "Invalid name.",
				NoSpace:     "No space left on the MULTITOOL partition.",
				Exists:      "A backup with this name already exists.",
				NoSource:    "No available source.",
			},
		}
	}
	return Catalog{
		ProductName: "GoMultitool",
		Menu: [7]MenuItem{
			{MenuBurn, "Gravar imagem no eMMC", "Instalar Armbian ou outra imagem"},
			{MenuBackup, "Fazer backup", "Salvar o eMMC na partição MULTITOOL"},
			{
				MenuHardware,
				"Informações de hardware",
				"Ver dispositivos de armazenamento conectados",
			},
			{MenuLegacy, "Menu antigo (Bash)", "Restaurar, apagar, steP-nand"},
			{MenuReboot, "Reiniciar", "Reiniciar o aparelho"},
			{MenuPowerOff, "Desligar", "Desligar o aparelho"},
			{MenuExit, "Sair", "Fechar o GoMultitool"},
		},
		HomeTitle:        "Menu de recuperação",
		HomeNav:          "↑/↓ ou ←/→ navegar  ·  Enter abrir  ·  q sair",
		FormNav:          "Tab/Shift+Tab ou ←/→ foco  ·  ↑/↓ navegar  ·  Enter selecionar  ·  Esc voltar",
		HardwareNav:      "↑/↓ rolar  ·  Esc voltar  ·  q sair",
		HardwareTitle:    "Dispositivos de armazenamento",
		EmptyInventory:   "Nenhum dispositivo encontrado.",
		ErrorPrefix:      "Erro: ",
		Loading:          "Carregando…",
		Busy:             "Outra operação em andamento.",
		ImageTooLarge:    "A imagem é maior que o destino.",
		DeviceChanged:    "O dispositivo mudou depois de escolhido; escolha de novo.",
		UnsafeDevice:     "Este dispositivo não pode ser gravado.",
		ReadbackMismatch: "A conferência falhou: os dados gravados diferem da imagem. Grave de novo.",
		NotSelectable:    "não selecionável",
		Unavailable:      "indisponível",
		OK:               "OK",
		ConfirmReboot:    "Reiniciar o aparelho agora?",
		ConfirmPowerOff:  "Desligar o aparelho agora?",
		PowerUnavailable: "Os controles de energia estão disponíveis na imagem de recuperação do GoMultitool.",
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
			Title:         "Gravar imagem",
			Review:        "Confirmar gravação",
			PickBoth:      "Pressione Enter para selecionar o destino e a imagem.",
			PickDevice:    "Pressione Enter no dispositivo que será apagado.",
			PickImage:     "Pressione Enter para selecionar a imagem.",
			Size:          "Tamanho",
			Destination:   "Destino",
			Images:        "Imagem",
			SizeAtBurn:    "tamanho calculado ao gravar",
			Warning:       "Tudo em %s será apagado.",
			Start:         "Gravar",
			Cancel:        "Cancelar",
			Cancelled:     "Gravação cancelada. O dispositivo ficou incompleto; grave de novo antes de usar.",
			Success:       "Gravado e conferido. Pode desligar.",
			CommandRate:   "A imagem não tem o command rate da DDR; o valor anterior não foi aplicado.",
			NoImages:      "Nenhuma imagem em MULTITOOL/images.",
			NoDestination: "Nenhum destino disponível.",
			Sizing:        "Calculando tamanho da imagem…",
			Verifying:     "Conferindo a gravação…",
		},
		Backup: BackupTexts{
			Title:       "Backup do dispositivo",
			PickSource:  "Pressione Enter para selecionar o dispositivo de origem.",
			Source:      "Origem",
			Name:        "Nome",
			Free:        "Livre na MULTITOOL: %s",
			Start:       "Fazer backup",
			Cancel:      "Cancelar",
			Cancelled:   "Backup cancelado. Nenhum arquivo foi salvo.",
			Saved:       "Backup salvo: %s.img.gz",
			InvalidName: "Nome inválido.",
			NoSpace:     "Sem espaço na partição MULTITOOL.",
			Exists:      "Já existe um backup com esse nome.",
			NoSource:    "Nenhuma origem disponível.",
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
