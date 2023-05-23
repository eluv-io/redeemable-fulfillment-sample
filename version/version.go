// Package version provides version information about the content fabric.
package version

import "time"

var date string
var full string

func init() {
	// if you see compile errors here, run `go generate version/gen.go`
	cd := time.Unix(commit_date, 0).UTC()
	date = cd.Format(time.RFC3339)
	if version != "" {
		full = version + " " + date
	} else if branch != "" {
		full = branch + "@" + revision + " " + date
	} else {
		full = `N/A - run 'go generate version/gen.go'`
	}
}

// Date returns the date of the last git commit in RFC3339 format.
func Date() string {
	return date
}

// Revision returns the revision hash of the last git commit.
func Revision() string {
	return revision
}

// Branch returns the git branch.
func Branch() string {
	return branch
}

// Version returns the tagged version. Is empty if the last commit was not
// tagged.
func Version() string {
	return version
}

// BestVersion returns the exact version, or, last tagged version if an exact is unavailable.
func BestVersion() string {
	if Version() == "" {
		return last_tag
	}
	return Version()
}

// Full returns the full version string.
//
// The format is [VERSION DATE] if a (tagged) version exists:
//     v1.0.0 2019-03-19T23:25:55+01:00
//
// Otherwise it is [BRANCH@REVISION DATE]:
//	 	add-version-info@a85787ff33cc94e001911a0368ca459aadc47eb9 2019-03-19T23:25:55+01:00
func Full() string {
	return full
}
