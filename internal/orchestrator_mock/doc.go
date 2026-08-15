// Package orchestrator_mock is a compileable sketch of APIHydra's target
// service boundaries. It deliberately owns mock copies of shared models so the
// proposed orchestration can be evaluated without making those contracts part
// of the production packages.
//
// The size annotations on services estimate production implementation lines;
// they do not describe the size of the mocks in this package. Tests and package
// documentation are excluded from those estimates. SMALL means fewer than
// roughly 200 lines, MEDIUM means roughly 200-600 lines, and BIG means roughly
// 600-1,000 lines.
package orchestrator_mock
