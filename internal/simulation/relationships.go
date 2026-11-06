package simulation

// A missing lock matrix keeps presets from earlier versions fully editable.
func (config Config) RelationshipLocked(row, column int) bool {
	return config.RelationshipLocks != nil && config.RelationshipLocks[row][column]
}

func (config *Config) ToggleRelationshipLock(row, column int) {
	if config.RelationshipLocks == nil {
		count := len(config.ParticleTypes)
		config.RelationshipLocks = make([][]bool, count)
		for index := range config.RelationshipLocks {
			config.RelationshipLocks[index] = make([]bool, count)
		}
	}
	config.RelationshipLocks[row][column] = !config.RelationshipLocks[row][column]
}

// SetRelationship is shared by manual, reciprocal, clear, and random edits.
func (config *Config) SetRelationship(row, column int, strength float64) {
	if !config.RelationshipLocked(row, column) {
		config.Relationships[row][column] = strength
	}
}
