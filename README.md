# closed-bugs

Searches a code base for references to closed bugs from
bugzilla.redhat.com.

## Examples

```
$ go run main.go ~/git/some-repo
Please wait, examining code base for closed bugs...
*** Bug 123456789 is CLOSED/ERRATA: some really important bug fix
  Link:
    https://bugzilla.redhat.com/show_bug.cgi?id=123456789
  References:
    ~/git/some-repo/test/unit.go:432
    ~/git/some-repo/rest/rest.go:19

exit status 1
```

```
$ go run main.go ~/git/some-repo
Please wait, examining code base for closed bugs...
Good job! No closed bugs found.
```
