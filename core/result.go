// Package core provides the public AgentResult type and related types for AgentFlow.
package core

import (
	"encoding/json"
	"time"
)

// AgentResult represents the result of an agent's execution.
type AgentResult struct {
	OutputState State     `json:"output_state"`
	Error       string    `json:"error,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    time.Duration
}

// MarshalJSON customizes the JSON encoding for AgentResult.
func (r *AgentResult) MarshalJSON() ([]byte, error) {
	durationMs := r.Duration.Milliseconds()
	jsonDataMap := map[string]interface{}{
		"output_state": r.OutputState,
		"error":        r.Error,
		"start_time":   r.StartTime,
		"end_time":     r.EndTime,
		"duration_ms":  durationMs,
	}
	if r.Error == "" {
		delete(jsonDataMap, "error")
	}
	jsonData, err := json.Marshal(jsonDataMap)
	if err != nil {
		Logger().Error().Err(err).Msg("MarshalJSON (map): Error during json.Marshal")
		return jsonData, err
	}
	return jsonData, nil
}

// UnmarshalJSON customizes the JSON decoding for AgentResult.
func (r *AgentResult) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		Logger().Error().Err(err).Msg("Unmarshal error (raw map)")
		return err
	}
	if rawOutputState, ok := raw["output_state"]; ok && string(rawOutputState) != "null" {
		var concreteState SimpleState
		if err := json.Unmarshal(rawOutputState, &concreteState); err != nil {
			Logger().Error().Err(err).Msg("Unmarshal error (output_state into SimpleState)")
			return err
		}
		r.OutputState = &concreteState
	} else {
		r.OutputState = nil
	}
	if rawError, ok := raw["error"]; ok {
		if err := json.Unmarshal(rawError, &r.Error); err != nil {
			Logger().Error().Err(err).Msg("Unmarshal error (error)")
			return err
		}
	}
	if rawStartTime, ok := raw["start_time"]; ok {
		if err := json.Unmarshal(rawStartTime, &r.StartTime); err != nil {
			Logger().Error().Err(err).Msg("Unmarshal error (start_time)")
			return err
		}
	}
	if rawEndTime, ok := raw["end_time"]; ok {
		if err := json.Unmarshal(rawEndTime, &r.EndTime); err != nil {
			Logger().Error().Err(err).Msg("Unmarshal error (end_time)")
			return err
		}
	}
	if rawDurationMs, ok := raw["duration_ms"]; ok {
		var durationMs int64
		if err := json.Unmarshal(rawDurationMs, &durationMs); err != nil {
			Logger().Error().Err(err).
				Str("raw", string(rawDurationMs)).
				Msg("Unmarshal error (duration_ms)")
			return err
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
	} else {
		Logger().Warn().Msg("UnmarshalJSON (manual): duration_ms field not found in JSON")
		r.Duration = 0
	}
	return nil
}
