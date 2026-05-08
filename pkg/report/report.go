package report

// TODO: Implement health_report tool
//
// This tool aggregates findings from all other tools and produces
// an overall "secret health score" (0-100) with prioritized recommendations.
//
// Key design decisions:
// - Define a scoring algorithm (weight by severity, normalize)
// - Call the other tools internally or accept their results as input?
//   Recommendation: accept results as input for better testability
// - Output a structured report with score, findings, and next steps
