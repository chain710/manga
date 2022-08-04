package db

func IndexVolumes(vols []Volume) map[string]Volume {
	m := make(map[string]Volume)
	for _, vol := range vols {
		m[vol.Path] = vol
	}
	return m
}
