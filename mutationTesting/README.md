# Mutation Testing

This Go example uses [Gremlins](https://github.com/go-gremlins/gremlins) to
check whether tests detect small changes to invoice rules.

An invoice is overdue when its balance is positive and its due date has passed.
A fixed 500-cent late fee applies only when a positive balance is more than
30 days overdue. This is a deliberately simplified example, not billing
software.

## Run

With Go (see `go.mod`) and Make installed, from this directory run:

```sh
make test-weak
make mutate-weak
make test
make mutate
```

The weak suite in `invoice/invoice_weak_test.go` exercises all statements but
misses boundary cases. `make test-weak` passes and reports coverage, while
`make mutate-weak` reports surviving mutations and lower test efficacy. A
surviving mutation does not by itself make the command exit with an error.
The corrected suite in
`invoice/invoice_valid_test.go` checks the boundaries and is selected by
default; `make test` and `make mutate` run it instead.

## What is wrong with the weak tests?

The weak tests cover both outcomes of each rule, but skip the exact values
where the rules change. As a result, they pass even when `>` is mistakenly
changed to `>=`:

| Rule | Missing case | Incorrect behavior that survives |
| --- | --- | --- |
| Overdue requires a positive balance | Balance is `0`, one day overdue | A paid invoice is marked overdue |
| Overdue requires a past due date | Positive balance, `0` days overdue | An invoice due today is marked overdue |
| Late fee requires a positive balance | Balance is `0`, 31 days overdue | A paid invoice is charged a late fee |
| Late fee applies after 30 days | Positive balance, 30 days overdue | A late fee is charged one day too early |

The corrected tests assert the expected result for each boundary. In this
example, both suites pass their normal tests and the weak suite reaches 100%
statement coverage, yet four boundary mutations survive the weak suite.

The suites are mutually exclusive: the `weak` Go build tag selects the weak
tests and excludes the corrected tests. From the repository root, add
`PROJECT=mutationTesting` to each command, for example
`make PROJECT=mutationTesting mutate-weak`.

The mutation targets run a pinned Gremlins version via `go run`, which downloads
it on first use; set `GREMLINS_VERSION` to select another version. No services
are required. They use one worker and a higher timeout coefficient so this small
test suite is less likely to time out when its coverage run is fast.

## Mutate changed code

For a pull request, fetch the base branch and enough history to find its merge
base, then run:

```sh
make mutate-diff BASE_REF=origin/main
# From the repository root:
make PROJECT=mutationTesting mutate-diff BASE_REF=origin/main
```

The target compares tracked changes in this module with the merge base of
`BASE_REF` and runs mutations in changed lines. Commit or stage new files
first: Git does not include untracked files in the diff. If there are no
tracked changes in the module, the target skips the run; Gremlins otherwise
treats an empty diff as a request to test all code.

The target uses module-relative paths and zero-context Git diffs to work around
Gremlins v0.6.0's nested-module path and changed-hunk parsing limitations. Test
this behavior when changing the Gremlins version. A change to tests alone may
produce no changed production mutations. Gremlins still collects coverage and
runs tests, so this is not limited to executing only changed tests. For tests
in other packages, consider Gremlins' `--integration` option; schedule broader
mutation runs separately.

## How to read the results

Gremlins changes operators and other expressions in the invoice code, then
reruns the tests against each mutant. A **killed** mutant causes a test failure;
a **lived** mutant does not. Not-covered and non-viable mutants have separate
statuses. A surviving mutation may indicate a missing assertion or an equivalent
change that cannot affect behavior; inspect it before changing the tests.

For example, changing `daysPastDue > 30` to `daysPastDue >= 30` incorrectly
charges a fee on day 30. The weak suite does not test day 30, so that mutation
survives. The corrected suite's `30 days late` test catches it. A passing
`go test` and high statement coverage alone do not prove that all incorrect
behavior is detected.
