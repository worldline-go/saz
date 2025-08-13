package database

type Batch struct {
	size int
	rows [][]any
}

func NewBatch(size int) *Batch {
	return &Batch{
		size: size,
		rows: make([][]any, 0, size),
	}
}

func (b *Batch) AddRow(row []any) {
	b.rows = append(b.rows, row)
}

func (b *Batch) IsFull() bool {
	return len(b.rows) >= b.size
}

func (b *Batch) Reset() {
	b.rows = b.rows[:0]
}

func (b *Batch) Size() int {
	return len(b.rows)
}

func (b *Batch) Rows() []any {
	if len(b.rows) == 0 {
		return nil
	}

	rows := make([]any, 0, len(b.rows)*len(b.rows[0]))
	for _, row := range b.rows {
		rows = append(rows, row...)
	}

	return rows
}
