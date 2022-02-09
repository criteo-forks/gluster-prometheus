package glusterutils

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Quotas returns gluster quotas (glusterd)
func (g *GD1) Quotas() ([]Quota, error) {
	out, err := g.execGluster("volume", "list")
	if err != nil {
		return nil, err
	}
	var vols gd1VolumesList
	if err := xml.Unmarshal(out, &vols); err != nil {
		return nil, err
	}

	var outQuotas []Quota
	for _, vol := range vols.List {
		out, err := g.execGluster("volume", "quota", vol, "list")
		if err != nil {
			return nil, err
		}

		var quotas gd1VolumeQuotas
		err = xml.Unmarshal(out, &quotas)
		if err != nil {
			return nil, err
		}

		for _, quota := range quotas.List {
			slp, err := percentStrToInt(quota.SoftLimitPercent)
			if err != nil {
				return nil, fmt.Errorf("failed to parse soft limit percent %v: %v", quota.SoftLimitPercent, err)
			}

			outq := Quota{
				Path:              quota.Path,
				Available:         quota.AvailSpace,
				Used:              quota.UsedSpace,
				SoftLimit:         quota.SoftLimitValue,
				SoftLimitPercent:  slp,
				SoftLimitExceeded: quota.SoftLimitExceeded == "Yes",
				HardLimit:         quota.HardLimit,
				HardLimitExceeded: quota.HardLimitExceeded == "Yes",
			}
			outQuotas = append(outQuotas, outq)
		}
	}
	return outQuotas, nil
}

func percentStrToInt(s string) (int, error) {
	s = strings.TrimSuffix(s, "%")
	return strconv.Atoi(s)
}
