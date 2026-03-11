package cmd

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const defaultBatchParallelism = "4"

func normalizeCLIArgs(args []string) []string {
	normalized := append([]string(nil), args...)
	for index, arg := range normalized {
		if strings.HasPrefix(arg, "-") {
			continue
		}

		if arg == "batch" {
			prefix := append([]string(nil), normalized[:index+1]...)
			suffix := normalizeBatchArgs(normalized[index+1:])
			return append(prefix, suffix...)
		}

		return normalized
	}

	return normalized
}

func normalizeArgsForCommand(command *cobra.Command, args []string) []string {
	normalized := append([]string(nil), args...)
	if command.Name() == "batch" {
		return normalizeBatchArgs(normalized)
	}

	return normalized
}

func normalizeBatchArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+1)

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg != "--parallel" {
			normalized = append(normalized, arg)
			continue
		}

		if index+1 < len(args) && looksLikeInt(args[index+1]) {
			normalized = append(normalized, "--parallel="+args[index+1])
			index++
			continue
		}

		normalized = append(normalized, "--parallel="+defaultBatchParallelism)
	}

	return normalized
}

func looksLikeInt(value string) bool {
	if strings.HasPrefix(value, "--") {
		return false
	}

	_, err := strconv.Atoi(value)
	return err == nil
}
