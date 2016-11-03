package core

// Analyzer interface that expresses the Analyzer's functionality
type Analyzer interface {
	// executes the analysis - blocking process
	Analyze() bool
	// returns the status of the Analyzer
	Status() AnalyzerStatus
	// returns the result of the analyzer - a serialized version of it
	Result() Result
	//returns the duration of the analysis
	Duration() float64
	// returns the dataset of the analysis
	Dataset() Dataset
}

// Status is the type representing the domain of the Analyzer's status
type AnalyzerStatus uint8

// Values of the AnalyzerStatus type
const (
	PENDING AnalyzerStatus = iota + 1
	ANALYZING
	ANALYZED
)

// String method is used to print the Status enum in a pretty manner
func (s AnalyzerStatus) String() string {
	switch s {
	case PENDING:
		return "PENDING"
	case ANALYZING:
		return "ANALYZING"
	case ANALYZED:
		return "ANALYZED"
	}
	return "UNKNOWN"
}

// Result struct holds the results of the analysis
type Result []float64
