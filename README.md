# well-locations

Testing serving GeoJSON using Go (using the chi router, jmoiron/sqlx, paulmach/orb)

Test dataset: 100k records, each with a coordinate location, resembling data from https://apps.nrs.gov.bc.ca/gwells

These endpoints require the following table:

```sql
CREATE TABLE well (
  well_tag_number serial primary key,
  geom Geometry(POINT)
);
```

## Methods:

Retrieve rows using `sqlx.Select`, iterate and create the GeoJSON features one by one, then return FeatureCollection response.

Time:  ~ 3s (`2.873050799s`, `3.824080601s`, `3.795641493s`, `2.522710575s`, `2.593235702s`, `2.565721455s`)

Quick hacky in memory cache: ~350 ms.
