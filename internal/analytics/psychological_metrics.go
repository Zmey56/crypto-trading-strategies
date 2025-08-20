package analytics

type UserProfile struct {
	ExperienceLevel string `json:"experience_level"`
	RiskTolerance   string `json:"risk_tolerance"`
	TimeHorizon     string `json:"time_horizon"`
}

type PsychologicalMetrics struct {
	Strategy            string  `json:"strategy"`
	StressLevel         float64 `json:"stress_level"`       // 1-10 scale
	MonitoringFrequency int     `json:"monitoring_per_day"` // times per day
	DecisionFatigue     float64 `json:"decision_fatigue"`   // 1-10 scale
	SleepQualityImpact  float64 `json:"sleep_impact"`       // 1-10 scale
	OverallSatisfaction float64 `json:"satisfaction"`       // 1-10 scale
}

// EvaluatePsychologicalImpact estimates the psychological impact of a strategy
func EvaluatePsychologicalImpact(strategy string, userProfile UserProfile) *PsychologicalMetrics {
	var stress, monitoring, fatigue, sleepImpact, satisfaction float64

	switch strategy {
	case "DCA":
		stress = 2.5
		monitoring = 1
		fatigue = 1.5
		sleepImpact = 1.2
		satisfaction = 8.5

	case "GRID":
		stress = 6.5
		monitoring = 8
		fatigue = 7.0
		sleepImpact = 5.5
		satisfaction = 7.2
	}

	// Adjust based on user profile
	if userProfile.ExperienceLevel == "BEGINNER" {
		stress += 2.0
		fatigue += 1.5
	}

	return &PsychologicalMetrics{
		Strategy:            strategy,
		StressLevel:         stress,
		MonitoringFrequency: int(monitoring),
		DecisionFatigue:     fatigue,
		SleepQualityImpact:  sleepImpact,
		OverallSatisfaction: satisfaction,
	}
}
