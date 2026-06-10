package plugins

// selfPackages are the names under which Pitara itself may appear in a global
// package manager's listing (e.g. once distributed via npm). Global-package
// plugins MUST skip these during Scan so a backup never records Pitara — on
// restore that would be redundant (Pitara must already be installed to run the
// restore) and could overwrite the running binary with a different version.
//
// Add every name/registry alias Pitara is published under as distribution grows.
var selfPackages = map[string]bool{
	"pitara":             true,
	"@sailingsam/pitara": true,
}

// IsSelf reports whether a global package name refers to Pitara itself.
func IsSelf(name string) bool {
	return selfPackages[name]
}
