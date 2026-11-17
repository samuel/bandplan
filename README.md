bandplan
========

`samuelks.com/bandplan` describes radio channel plans: band sets made of named
bands, each a list of channels with a center frequency and width. It has no
dependencies.

- Built-in sets: 2.4 and 5 GHz WiFi, FPV video (5.8 GHz), Zigbee, Bluetooth,
  LoRaWAN US915 and EU868, LPD433, PMR446, MURS, GMRS, marine VHF, NOAA
  Weather Radio, and US amateur VHF/UHF. `Builtins` and `Builtin(name)`
  return copies.
- `LoadFile`/`SaveFile` (and `Decode`/`Encode`) read and write sets as JSON,
  and `MergeSets` lets custom sets replace built-ins of the same name.
- `BandSet.Span` gives the lowest and highest channel edge, and
  `BandSet.Lookup(f)` the channels containing a frequency.

JSON
----

[`bandplan.json`](bandplan.json) holds the built-in sets for use outside Go.
It is generated with `go generate ./...`, and a test fails when it is out of
date. [`bandplan.schema.json`](bandplan.schema.json) (JSON Schema 2020-12)
describes the format, which is the same for user files:

```json
{
  "version": 1,
  "sets": [
    {
      "name": "My radios",
      "bands": [
        {
          "name": "UHF",
          "channels": [{"name": "1", "centerHz": 462562500, "widthHz": 12500}]
        }
      ]
    }
  ]
}
```

Frequencies are integer hertz. They may not exceed 2^53-1
(9007199254740991), the largest integer an IEEE 754 double holds exactly, so
JavaScript and other decoders that read numbers as doubles get exact values.
A channel spans `centerHz - widthHz/2` to `centerHz + widthHz/2`, edges
inclusive. Every set, band, and channel needs a name, and centers and widths
must be positive. Channels within a set may overlap. `version` changes only
when an existing reader would misread a file; readers should reject versions
they don't know, and unknown fields are rejected.
