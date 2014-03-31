/*
The stenographer is used by Ginkgo's reporters to generate output.

Move along, nothing to see here.
*/

package stenographer

import (
	"fmt"
	"github.com/onsi/ginkgo/types"
	"strings"
)

const defaultStyle = "\x1b[0m"
const boldStyle = "\x1b[1m"
const redColor = "\x1b[91m"
const greenColor = "\x1b[32m"
const yellowColor = "\x1b[33m"
const cyanColor = "\x1b[36m"
const grayColor = "\x1b[90m"
const lightGrayColor = "\x1b[37m"

type cursorStateType int

const (
	cursorStateTop cursorStateType = iota
	cursorStateStreaming
	cursorStateMidBlock
	cursorStateEndBlock
)

type Stenographer interface {
	AnnounceSuite(description string, randomSeed int64, randomizingAll bool, succinct bool)
	AnnounceAggregatedParallelRun(nodes int, succinct bool)
	AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int, succinct bool)
	AnnounceNumberOfSpecs(specsToRun int, total int, succinct bool)
	AnnounceSpecRunCompletion(summary *types.SuiteSummary, succinct bool)

	AnnounceSpecWillRun(spec *types.SpecSummary)

	AnnounceCapturedOutput(spec *types.SpecSummary)

	AnnounceSuccesfulSpec(spec *types.SpecSummary)
	AnnounceSuccesfulSlowSpec(spec *types.SpecSummary, succinct bool)
	AnnounceSuccesfulMeasurement(spec *types.SpecSummary, succinct bool)

	AnnouncePendingSpec(spec *types.SpecSummary, noisy bool)
	AnnounceSkippedSpec(spec *types.SpecSummary)

	AnnounceSpecTimedOut(spec *types.SpecSummary, succinct bool)
	AnnounceSpecPanicked(spec *types.SpecSummary, succinct bool)
	AnnounceSpecFailed(spec *types.SpecSummary, succinct bool)
}

func New(color bool) Stenographer {
	return &consoleStenographer{
		color:       color,
		cursorState: cursorStateTop,
	}
}

type consoleStenographer struct {
	color       bool
	cursorState cursorStateType
}

var alternatingColors = []string{defaultStyle, grayColor}

func (s *consoleStenographer) AnnounceSuite(description string, randomSeed int64, randomizingAll bool, succinct bool) {
	if succinct {
		s.print(0, "[%d] %s ", randomSeed, s.colorize(boldStyle, description))
		return
	}
	s.printBanner(fmt.Sprintf("Running Suite: %s", description), "=")
	s.print(0, "Random Seed: %s", s.colorize(boldStyle, "%d", randomSeed))
	if randomizingAll {
		s.print(0, " - Will randomize all specs")
	}
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int, succinct bool) {
	if succinct {
		s.print(0, "- node #%d ", node)
		return
	}
	s.println(0,
		"Parallel test node %s/%s. Assigned %s of %s specs.",
		s.colorize(boldStyle, "%d", node),
		s.colorize(boldStyle, "%d", nodes),
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", totalSpecs),
	)
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceAggregatedParallelRun(nodes int, succinct bool) {
	if succinct {
		s.print(0, "- %d nodes ", nodes)
		return
	}
	s.println(0,
		"Running in parallel across %s nodes",
		s.colorize(boldStyle, "%d", nodes),
	)
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceNumberOfSpecs(specsToRun int, total int, succinct bool) {
	if succinct {
		s.print(0, "- %d/%d specs ", specsToRun, total)
		s.cursorState = cursorStateStreaming
		return
	}
	s.println(0,
		"Will run %s of %s specs",
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", total),
	)

	s.printNewLine()
}

func (s *consoleStenographer) AnnounceSpecRunCompletion(summary *types.SuiteSummary, succinct bool) {
	if succinct && summary.SuiteSucceeded {
		s.print(0, " %s %s ", s.colorize(greenColor, "SUCCESS!"), summary.RunTime)
		return
	}
	s.printNewLine()
	color := greenColor
	if !summary.SuiteSucceeded {
		color = redColor
	}
	s.println(0, s.colorize(boldStyle+color, "Ran %d of %d Specs in %.3f seconds", summary.NumberOfSpecsThatWillBeRun, summary.NumberOfTotalSpecs, summary.RunTime.Seconds()))

	status := ""
	if summary.SuiteSucceeded {
		status = s.colorize(boldStyle+greenColor, "SUCCESS!")
	} else {
		status = s.colorize(boldStyle+redColor, "FAIL!")
	}

	s.print(0,
		"%s -- %s | %s | %s | %s ",
		status,
		s.colorize(greenColor+boldStyle, "%d Passed", summary.NumberOfPassedSpecs),
		s.colorize(redColor+boldStyle, "%d Failed", summary.NumberOfFailedSpecs),
		s.colorize(yellowColor+boldStyle, "%d Pending", summary.NumberOfPendingSpecs),
		s.colorize(cyanColor+boldStyle, "%d Skipped", summary.NumberOfSkippedSpecs),
	)
}

func (s *consoleStenographer) AnnounceSpecWillRun(spec *types.SpecSummary) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	}
	for i, text := range spec.ComponentTexts[1 : len(spec.ComponentTexts)-1] {
		s.print(0, s.colorize(alternatingColors[i%2], text)+" ")
	}

	indentation := 0
	if len(spec.ComponentTexts) > 2 {
		indentation = 1
		s.printNewLine()
	}
	index := len(spec.ComponentTexts) - 1
	s.print(indentation, s.colorize(boldStyle, spec.ComponentTexts[index]))
	s.printNewLine()
	s.print(indentation, s.colorize(lightGrayColor, spec.ComponentCodeLocations[index].String()))
	s.printNewLine()
	s.cursorState = cursorStateMidBlock
}

func (s *consoleStenographer) AnnounceCapturedOutput(spec *types.SpecSummary) {
	if spec.CapturedOutput == "" {
		return
	}

	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}
	s.println(0, spec.CapturedOutput)
	s.cursorState = cursorStateMidBlock
}

func (s *consoleStenographer) AnnounceSuccesfulSpec(spec *types.SpecSummary) {
	s.print(0, s.colorize(greenColor, "•"))
	s.cursorState = cursorStateStreaming
}

func (s *consoleStenographer) AnnounceSuccesfulSlowSpec(spec *types.SpecSummary, succinct bool) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [SLOW TEST:%.3f seconds]", spec.RunTime.Seconds()),
		"",
		spec,
		succinct,
	)
}

func (s *consoleStenographer) AnnounceSuccesfulMeasurement(spec *types.SpecSummary, succinct bool) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [MEASUREMENT]"),
		s.measurementReport(spec),
		spec,
		succinct,
	)
}

func (s *consoleStenographer) AnnouncePendingSpec(spec *types.SpecSummary, noisy bool) {
	if noisy {
		s.printBlockWithMessage(
			s.colorize(yellowColor, "P [PENDING]"),
			"",
			spec,
			false,
		)
	} else {
		s.print(0, s.colorize(greenColor, "P"))
		s.cursorState = cursorStateStreaming
	}
}

func (s *consoleStenographer) AnnounceSkippedSpec(spec *types.SpecSummary) {
	s.print(0, s.colorize(cyanColor, "S"))
	s.cursorState = cursorStateStreaming
}

func (s *consoleStenographer) AnnounceSpecTimedOut(spec *types.SpecSummary, succinct bool) {
	s.printFailure("•... Timeout", spec, succinct)
}

func (s *consoleStenographer) AnnounceSpecPanicked(spec *types.SpecSummary, succinct bool) {
	s.printFailure("•! Panic", spec, succinct)
}

func (s *consoleStenographer) AnnounceSpecFailed(spec *types.SpecSummary, succinct bool) {
	s.printFailure("• Failure", spec, succinct)
}

func (s *consoleStenographer) printBlockWithMessage(header string, message string, spec *types.SpecSummary, succinct bool) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, header)

	indentation := s.printCodeLocationBlock(spec, false, succinct)

	if message != "" {
		s.printNewLine()
		s.println(indentation, message)
	}

	s.printDelimiter()
	s.cursorState = cursorStateEndBlock
}

func (s *consoleStenographer) printFailure(message string, spec *types.SpecSummary, succinct bool) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, s.colorize(redColor+boldStyle, "%s [%.3f seconds]", message, spec.RunTime.Seconds()))

	indentation := s.printCodeLocationBlock(spec, true, succinct)

	s.printNewLine()
	if spec.State == types.SpecStatePanicked {
		s.println(indentation, s.colorize(redColor+boldStyle, spec.Failure.Message))
		s.println(indentation, s.colorize(redColor, "%v", spec.Failure.ForwardedPanic))
		s.println(indentation, spec.Failure.Location.String())
		s.printNewLine()
		s.println(indentation, s.colorize(redColor, "Full Stack Trace"))
		s.println(indentation, spec.Failure.Location.FullStackTrace)
	} else {
		s.println(indentation, s.colorize(redColor, spec.Failure.Message))
		s.printNewLine()
		s.println(indentation, spec.Failure.Location.String())
	}

	s.printDelimiter()
	s.cursorState = cursorStateEndBlock
}

func (s *consoleStenographer) printCodeLocationBlock(spec *types.SpecSummary, failure bool, succinct bool) int {
	indentation := 0
	startIndex := 1

	if len(spec.ComponentTexts) == 1 {
		startIndex = 0
	}

	for i := startIndex; i < len(spec.ComponentTexts); i++ {
		if failure && i == spec.Failure.ComponentIndex {
			blockType := ""
			switch spec.Failure.ComponentType {
			case types.SpecComponentTypeBeforeEach:
				blockType = "BeforeEach"
			case types.SpecComponentTypeJustBeforeEach:
				blockType = "JustBeforeEach"
			case types.SpecComponentTypeAfterEach:
				blockType = "AfterEach"
			case types.SpecComponentTypeIt:
				blockType = "It"
			case types.SpecComponentTypeMeasure:
				blockType = "Measurement"
			}
			if succinct {
				s.print(0, s.colorize(redColor+boldStyle, "[%s] %s ", blockType, spec.ComponentTexts[i]))
			} else {
				s.println(indentation, s.colorize(redColor+boldStyle, "%s [%s]", spec.ComponentTexts[i], blockType))
				s.println(indentation, s.colorize(grayColor, "(%s)", spec.ComponentCodeLocations[i]))
			}
		} else {
			if succinct {
				s.print(0, s.colorize(alternatingColors[i%2], "%s ", spec.ComponentTexts[i]))
			} else {
				s.println(indentation, spec.ComponentTexts[i])
				s.println(indentation, s.colorize(grayColor, "(%s)", spec.ComponentCodeLocations[i]))
			}
		}

		indentation++
	}

	if succinct {
		if len(spec.ComponentTexts) > 0 {
			s.printNewLine()
			s.print(0, s.colorize(lightGrayColor, "(%s)", spec.ComponentCodeLocations[len(spec.ComponentCodeLocations)-1]))
		}
		s.printNewLine()
		indentation = 1
	} else {
		indentation--
	}

	return indentation
}

func (s *consoleStenographer) measurementReport(spec *types.SpecSummary) string {
	if len(spec.Measurements) == 0 {
		return "Found no measurements"
	}

	message := []string{}

	message = append(message, fmt.Sprintf("Ran %s samples:", s.colorize(boldStyle, "%d", spec.NumberOfSamples)))
	i := 0
	for _, measurement := range spec.Measurements {
		if i > 0 {
			message = append(message, "\n")
		}
		info := ""
		if measurement.Info != nil {
			message = append(message, fmt.Sprintf("%v", measurement.Info))
		}

		message = append(message, fmt.Sprintf("%s:\n%s  %s: %s%s\n  %s: %s%s\n  %s: %s%s ± %s%s",
			s.colorize(boldStyle, "%s", measurement.Name),
			info,
			measurement.SmallestLabel,
			s.colorize(greenColor, "%.3f", measurement.Smallest),
			measurement.Units,
			measurement.LargestLabel,
			s.colorize(redColor, "%.3f", measurement.Largest),
			measurement.Units,
			measurement.AverageLabel,
			s.colorize(cyanColor, "%.3f", measurement.Average),
			measurement.Units,
			s.colorize(cyanColor, "%.3f", measurement.StdDeviation),
			measurement.Units,
		))
		i++
	}

	return strings.Join(message, "\n")
}