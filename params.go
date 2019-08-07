// Params loads and holds configuration for runtime
package main

// Parameters holds the application's parameters it runs on
type Parameters struct {

	// Raw data parameters
	Filter			string	// BPF filter to filter traffic sniffing
	CollectorFile	string 	// File the Collector dumps data in, and the Monitor reads from

	// Analysis related parameters
	ProbePeriod			int	// Time (seconds) between analyses of collected data
	AlertSpan			int	// Time (seconds) span to monitor for alert trigger
	AlertThreshold		int	// Traffic (hits/span) threshold triggering an alert

	// Display related parameters
	DisplayRefresh	int		// Time (seconds) the display will be updated
	DisplayFormat	string	// Format of result of allowedFormats
	Output			string	// Output destinations among allowedOutputs
}

// Default values for Parameter object
const (
	defFilter			=	"tcp and port 80"
	defCollectorFile	=	"./gonetmon.dump"
	defProbePeriod 		=	1
	defAlertSpan		=	120
	defAlertThreshold	=	500
	defDisplayRefresh	=	10
	defDisplayFormat	=	"plain"
	defOutput			=	"cli"
)

// LoadParams loads the application's parameters it should run on into an object and returns it
func LoadParams() *Parameters{
	// Todo : There should be a better way of doing this + argument validation

	return &Parameters{
		CollectorFile:	defCollectorFile,
		ProbePeriod:	defProbePeriod,
		AlertSpan:		defAlertSpan,
		AlertThreshold:	defAlertThreshold,
		DisplayRefresh:	defDisplayRefresh,
		DisplayFormat:	defDisplayFormat,
		Output: 		defOutput,
	}
}