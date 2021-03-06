/*
Package s2 encapsulates Google's S2 library.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki

...and in the Google S2 documentation:

https://s2geometry.io/

*/
package s2

import (
	gs2 "github.com/golang/geo/s2"
)

// Default cell level, see https://s2geometry.io/resources/s2cell_statistics.html.
var DefaultLevel = 21

// Token returns the S2 cell token for coordinates using the default level.
func Token(lat, lng float64) string {
	return TokenLevel(lat, lng, DefaultLevel)
}

// Token returns the S2 cell token for coordinates.
func TokenLevel(lat, lng float64, level int) string {
	if lat == 0.0 && lng == 0.0 {
		return ""
	}

	if lat < -90 || lat > 90 {
		return ""
	}

	if lng < -180 || lng > 180 {
		return ""
	}

	l := gs2.LatLngFromDegrees(lat, lng)
	return gs2.CellIDFromLatLng(l).Parent(level).ToToken()
}

// LatLng returns the coordinates for a S2 cell token.
func LatLng(token string) (lat, lng float64) {
	if token == "" || token == "-" {
		return 0.0, 0.0
	}

	c := gs2.CellIDFromToken(token)
	l := c.LatLng()
	return l.Lat.Degrees(), l.Lng.Degrees()
}
