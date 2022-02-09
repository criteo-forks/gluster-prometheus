package glusterutils

import "errors"

// Quotas returns gluster quotas
func (g *GD2) Quotas() ([]Quota, error) {
	return nil, errors.New("not implemented")
}
