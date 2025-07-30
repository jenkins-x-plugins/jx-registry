package rootcmd

import "os"

var (
	// TopLevelCommand top level command
	TopLevelCommand = "jx-registry"

	// BinaryName the name of the command binary in help
	BinaryName = "jx-registry"
)

func init() {
	binaryName := os.Getenv("BINARY_NAME")
	if binaryName != "" {
		BinaryName = binaryName
	}
	topLevelCommand := os.Getenv("TOP_LEVEL_COMMAND")
	if topLevelCommand != "" {
		TopLevelCommand = topLevelCommand
	}
}
