package wireguard

import "strings"

// ReplaceOrAppendManagedBlock replaces the managed block between markers if present,
// otherwise appends it to the end of the file (with a separating newline).
func ReplaceOrAppendManagedBlock(original string, managedBlock string) string {
	orig := original
	if strings.TrimSpace(managedBlock) == "" {
		return orig
	}

	// Normalize: ensure managed block ends with newline for clean file layout
	if !strings.HasSuffix(managedBlock, "\n") {
		managedBlock += "\n"
	}

	beginIdx := strings.Index(orig, ManagedBlockBegin)
	endIdx := strings.Index(orig, ManagedBlockEnd)
	if beginIdx == -1 || endIdx == -1 || endIdx < beginIdx {
		trimmed := strings.TrimRight(orig, "\n")
		if trimmed == "" {
			return managedBlock
		}
		return trimmed + "\n\n" + managedBlock
	}

	// Replace from begin line start to end marker line end.
	start := beginIdx
	if lineStart := strings.LastIndex(orig[:beginIdx], "\n"); lineStart != -1 {
		start = lineStart + 1
	}
	afterEnd := endIdx + len(ManagedBlockEnd)
	if afterEnd < len(orig) {
		if nl := strings.Index(orig[afterEnd:], "\n"); nl != -1 {
			afterEnd = afterEnd + nl + 1
		} else {
			afterEnd = len(orig)
		}
	}

	return orig[:start] + managedBlock + orig[afterEnd:]
}
