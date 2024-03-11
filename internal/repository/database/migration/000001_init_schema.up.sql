CREATE TABLE gauge_metrics (
		m_name TEXT UNIQUE,
		m_value DOUBLE PRECISION
	);
CREATE TABLE counter_metrics (
		m_name TEXT UNIQUE,
		m_value INTEGER
	);
