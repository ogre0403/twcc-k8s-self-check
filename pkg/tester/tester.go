package tester

const (
	PASS = "PASS"
	FAIL = "FAIL"
)

type Tester interface {
	// Run Test case
	Run(req interface{}) Tester

	// Check Test result
	Check() Tester

	// fill report
	Report(interface{}) Tester

	// if need to run next step
	Next() bool

	// close opened resource by tester
	Close()
}
