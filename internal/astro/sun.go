package astro

import (
	"math"
	"time"
)

// SunTimes returns the sunrise and sunset times for the given date and location
// using the NOAA Solar Calculator algorithm. Returns zero times if lat/lon are
// both zero (location not configured) or if the location experiences midnight sun
// or polar night on the given date.
func SunTimes(t time.Time, lat, lon float64) (rise, set time.Time) {
	if lat == 0 && lon == 0 {
		return
	}

	local := t.Local()
	_, utcOffsetSec := local.Zone()
	utcOffsetHrs := float64(utcOffsetSec) / 3600.0

	jd := julianDay(local.Year(), int(local.Month()), local.Day())
	T := (jd - 2451545.0) / 36525.0

	// Geometric mean longitude and anomaly of the sun (degrees).
	L0 := math.Mod(280.46646+T*(36000.76983+T*0.0003032), 360.0)
	M := 357.52911 + T*(35999.05029-T*0.0001537)

	// Equation of center → true longitude → apparent longitude.
	Mrad := sunRad(M)
	C := (1.914602-T*(0.004817+T*0.000014))*math.Sin(Mrad) +
		(0.019993-T*0.000101)*math.Sin(2*Mrad) +
		0.000289*math.Sin(3*Mrad)
	omega := 125.04 - 1934.136*T
	lambda := L0 + C - 0.00569 - 0.00478*math.Sin(sunRad(omega))

	// Sun's declination.
	eps := 23.0 + (26.0+(21.448-T*(46.815+T*(0.00059-T*0.001813)))/60.0)/60.0
	eps += 0.00256 * math.Cos(sunRad(omega))
	decl := math.Asin(math.Sin(sunRad(eps)) * math.Sin(sunRad(lambda)))

	// Equation of time (minutes).
	e := 0.016708634 - T*(0.000042037+T*0.0000001267)
	y := math.Pow(math.Tan(sunRad(eps/2)), 2)
	L0rad := sunRad(L0)
	eqT := 4 * sunDeg(y*math.Sin(2*L0rad)-
		2*e*math.Sin(Mrad)+
		4*e*y*math.Sin(Mrad)*math.Cos(2*L0rad)-
		0.5*y*y*math.Sin(4*L0rad)-
		1.25*e*e*math.Sin(2*Mrad))

	// Hour angle at sunrise (positive means sunrise exists).
	latRad := sunRad(lat)
	cosHA := math.Cos(sunRad(90.833))/(math.Cos(latRad)*math.Cos(decl)) - math.Tan(latRad)*math.Tan(decl)
	if cosHA < -1 || cosHA > 1 {
		return // midnight sun or polar night
	}
	ha := sunDeg(math.Acos(cosHA))

	// Solar noon in minutes past local midnight.
	solarNoon := 720 - 4*lon - eqT + utcOffsetHrs*60

	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	rise = midnight.Add(time.Duration((solarNoon - 4*ha) * float64(time.Minute)))
	set = midnight.Add(time.Duration((solarNoon + 4*ha) * float64(time.Minute)))
	return
}

// julianDay returns the Julian Day Number for noon UTC on the given date.
func julianDay(year, month, day int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	A := math.Floor(float64(year) / 100)
	B := 2 - A + math.Floor(A/4)
	return math.Floor(365.25*float64(year+4716)) +
		math.Floor(30.6001*float64(month+1)) +
		float64(day) + B - 1524.5
}

func sunRad(deg float64) float64 { return deg * math.Pi / 180 }
func sunDeg(rad float64) float64 { return rad * 180 / math.Pi }
