# Golden charts

One reference chart per file, cross-checked against 2–3 public calculators.
Every engine implementation must pass all of them (`go test ./...`).
Format: see `_template.json` (files starting with `_` are not run).

Required coverage (target 20–30 files):

- [ ] every type: generator, manifesting_generator, projector, manifestor, reflector
- [ ] every authority: emotional, sacral, splenic, ego_manifested, ego_projected, self_projected, mental, lunar
- [ ] a planet within ±0.01° of a gate boundary, and of a line boundary
- [ ] birth during a DST jump: nonexistent local time and ambiguous local time
- [ ] USSR/RF zones: decree time, 1981 DST start, 1991 transition, 2011 (permanent "summer"), 2014 (back to winter), Samara and Kaliningrad switches
- [ ] births before 1950 and after 2000
- [ ] southern hemisphere and a western-longitude zone (America/*)
