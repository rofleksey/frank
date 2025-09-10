package act

import "strings"

func generateDescription(cmds []Command) string {
	var builder strings.Builder

	for _, cmd := range cmds {
		builder.WriteString("## ")
		builder.WriteString(cmd.Name())
		builder.WriteString(" COMMAND\n")

		builder.WriteString(cmd.Description())
		builder.WriteString("\n\n")
	}

	return builder.String()
}
