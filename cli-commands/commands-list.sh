# To check unique versions of a package accross all go.mod files in a repo
grep -R <package name> --include="go.mod" . | awk '{print $2, $3}' | sort | uniq -c