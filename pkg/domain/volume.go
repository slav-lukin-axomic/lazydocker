package domain

// Volume is the flat, framework-free model of a Docker volume. It replaces the
// pre-migration commands.Volume, which embedded the SDK's volume.Volume and
// carried a live client. Status stays a map[string]any so the config view's
// FormatMapItem call is unchanged and a nil Status still renders "n/a";
// UsageData stays a pointer so a nil one skips the RefCount/Size lines.
type Volume struct {
	Name       string
	Driver     string
	Scope      string
	Mountpoint string
	Labels     map[string]string
	Options    map[string]string
	Status     map[string]any
	UsageData  *VolumeUsageData
}

// VolumeUsageData holds the subset of the SDK's volume.UsageData the config view
// renders. It is present only when the Engine reported usage data.
type VolumeUsageData struct {
	RefCount int64
	Size     int64
}
