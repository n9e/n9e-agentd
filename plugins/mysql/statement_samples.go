package mysql

func NewMySQLStatementSamples() *MySQLStatementSamples {
	return &MySQLStatementSamples{}
}

type MySQLStatementSamples struct {
}

func (p *MySQLStatementSamples) Cancel() {
}

func (p *MySQLStatementSamples) runSampler() {
}
