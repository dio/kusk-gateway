package controllers

const (
	reconcilerDefaultRetrySeconds int = 30
	reconcilerFastRetrySeconds    int = 1
)

// Agent functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
