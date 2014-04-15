package apk

type Instrumentation struct {
	Name            string `xml:"name,attr"`
	Target          string `xml:"targetPackage,attr"`
	HandleProfiling bool   `xml:"handleProfiling,attr"`
	FunctionalTest  bool   `xml:"functionalTest,attr"`
}

type Application struct {
	AllowTaskReparenting  bool   `xml:"allowTaskReparenting,attr"`
	AllowBackup           bool   `xml:"allowBackup,attr"`
	BackupAgent           string `xml:"backupAgent,attr"`
	Debuggable            bool   `xml:"debuggable,attr"`
	Description           string `xml:"description,attr"`
	Enabled               bool   `xml:"enabled,attr"`
	HasCode               bool   `xml:"hasCode,attr"`
	HardwareAccelerated   bool   `xml:"hardwareAccelerated,attr"`
	Icon                  string `xml:"icon,attr"`
	KillAfterRestore      bool   `xml:"killAfterRestore,attr"`
	LargeHeap             bool   `xml:"largeHeap,attr"`
	Label                 string `xml:"label,attr"`
	Logo                  int    `xml:"logo,attr"`
	ManageSpaceActivity   string `xml:"manageSpaceActivity,attr"`
	Name                  string `xml:"name,attr"`
	Permission            string `xml:"permission,attr"`
	Persistent            bool   `xml:"persistent,attr"`
	Process               string `xml:"process,attr"`
	RestoreAnyVersion     bool   `xml:"restoreAnyVersion,attr"`
	RequiredAccountType   string `xml:"requiredAccountType,attr"`
	RestrictedAccountType string `xml:"restrictedAccountType,attr"`
	SupportsRtl           bool   `xml:"supportsRtl,attr"`
	TaskAffinity          string `xml:"taskAffinity,attr"`
	TestOnly              bool   `xml:"testOnly,attr"`
	Theme                 int    `xml:"theme,attr"`
	UiOptions             string `xml:"uiOptions,attr"`
	VmSafeMode            bool   `xml:"vmSafeMode,attr"`
}

type Manifest struct {
	Package     string          `xml:"package,attr"`
	VersionCode int             `xml:"versionCode,attr"`
	VersionName string          `xml:"versionName,attr"`
	App         Application     `xml:"application"`
	Instrument  Instrumentation `xml:"instrumentation"`
}
