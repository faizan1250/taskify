package config

func (c *CronConfig) ApplyDefaultsForTest(def *CronConfig) {
	c.applyDefaults(def)
}
