package plist

// LaunchAgentPlist represents a parsed macOS launch agent or daemon plist.
type LaunchAgentPlist struct {
	Label                     string                 `plist:"Label"`
	Disabled                  bool                   `plist:"Disabled,omitempty"`
	Program                   string                 `plist:"Program,omitempty"`
	ProgramArguments          []string               `plist:"ProgramArguments,omitempty"`
	EnableGlobbing            bool                   `plist:"EnableGlobbing,omitempty"`
	EnvironmentVariables      map[string]string      `plist:"EnvironmentVariables,omitempty"`
	WorkingDirectory          string                 `plist:"WorkingDirectory,omitempty"`
	StandardOutPath           string                 `plist:"StandardOutPath,omitempty"`
	StandardErrorPath         string                 `plist:"StandardErrorPath,omitempty"`
	StandardInPath            string                 `plist:"StandardInPath,omitempty"`
	RunAtLoad                 bool                   `plist:"RunAtLoad,omitempty"`
	KeepAlive                 interface{}            `plist:"KeepAlive,omitempty"`
	StartInterval             int                    `plist:"StartInterval,omitempty"`
	StartCalendarInterval     interface{}            `plist:"StartCalendarInterval,omitempty"`
	StartOnMount              bool                   `plist:"StartOnMount,omitempty"`
	WatchPaths                []string               `plist:"WatchPaths,omitempty"`
	QueueDirectories          []string               `plist:"QueueDirectories,omitempty"`
	UserName                  string                 `plist:"UserName,omitempty"`
	GroupName                 string                 `plist:"GroupName,omitempty"`
	Umask                     interface{}            `plist:"Umask,omitempty"`
	RootDirectory             string                 `plist:"RootDirectory,omitempty"`
	ExitTimeOut               int                    `plist:"ExitTimeOut,omitempty"`
	ThrottleInterval          int                    `plist:"ThrottleInterval,omitempty"`
	InitGroups                *bool                  `plist:"InitGroups,omitempty"`
	Nice                      int                    `plist:"Nice,omitempty"`
	ProcessType               string                 `plist:"ProcessType,omitempty"`
	AbandonProcessGroup       bool                   `plist:"AbandonProcessGroup,omitempty"`
	LowPriorityIO             bool                   `plist:"LowPriorityIO,omitempty"`
	LowPriorityBackgroundIO   bool                   `plist:"LowPriorityBackgroundIO,omitempty"`
	LaunchOnlyOnce            bool                   `plist:"LaunchOnlyOnce,omitempty"`
	MachServices              map[string]interface{} `plist:"MachServices,omitempty"`
	Sockets                   map[string]interface{} `plist:"Sockets,omitempty"`
	LaunchEvents              map[string]interface{} `plist:"LaunchEvents,omitempty"`
	HardResourceLimits        map[string]int         `plist:"HardResourceLimits,omitempty"`
	SoftResourceLimits        map[string]int         `plist:"SoftResourceLimits,omitempty"`
	EnableTransactions        bool                   `plist:"EnableTransactions,omitempty"`
	EnablePressuredExit       bool                   `plist:"EnablePressuredExit,omitempty"`
	Debug                     bool                   `plist:"Debug,omitempty"`
	WaitForDebugger           bool                   `plist:"WaitForDebugger,omitempty"`
	LimitLoadToSessionType    interface{}            `plist:"LimitLoadToSessionType,omitempty"`
	LimitLoadToHardware       map[string][]string    `plist:"LimitLoadToHardware,omitempty"`
	LimitLoadFromHardware     map[string][]string    `plist:"LimitLoadFromHardware,omitempty"`
	InetdCompatibility        map[string]bool        `plist:"inetdCompatibility,omitempty"`
	AssociatedBundleIdentifiers interface{}           `plist:"AssociatedBundleIdentifiers,omitempty"`
}

// ProgramPath returns the effective program path. It prefers Program,
// falling back to the first element of ProgramArguments.
func (p *LaunchAgentPlist) ProgramPath() string {
	if p.Program != "" {
		return p.Program
	}
	if len(p.ProgramArguments) > 0 {
		return p.ProgramArguments[0]
	}
	return ""
}

// ValidationError describes a problem found when validating a plist.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
