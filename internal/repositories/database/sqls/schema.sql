CREATE TABLE IF NOT EXISTS gauge_metrics (
		m_name TEXT UNIQUE,
		m_value DOUBLE PRECISION
	);
CREATE TABLE IF NOT EXISTS counter_metrics (
		m_name TEXT UNIQUE,
		m_value INTEGER
	);
