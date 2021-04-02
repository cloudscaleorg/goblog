package dtree

import (
	"context"
)

// Decision is a  node in a decision tree.
//
// Depending on if its Exec function returns
// true or false its yes or no branch will be
// followed respectively.
type Decision struct {
	yes  *Decision
	no   *Decision
	Exec func(ctx context.Context) (bool, error)
}

func (d *Decision) AddYes(node *Decision) {
	d.yes = node
}

func (d *Decision) AddNo(node *Decision) {
	d.no = node
}

// Execute recursively executes all children Exec nodes
// of this root Decision.
//
// This method will return the first Decision node that
// encounters an error, including itself.
func (d *Decision) Execute(ctx context.Context) error {
	b, err := d.Exec(ctx)
	if err != nil {
		return err
	}
	if b {
		if d.yes == nil {
			return nil
		}
		return d.yes.Execute(ctx)
	}
	if d.no == nil {
		return nil
	}
	return d.no.Execute(ctx)
}
