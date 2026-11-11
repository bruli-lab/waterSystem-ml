import json
import os
from dataclasses import dataclass
from datetime import date
from functools import lru_cache
from pathlib import Path

import pandas as pd

DEFAULT_ZONES_FILE = Path(__file__).resolve().parent.parent / "manifests" / "zones.json"


@dataclass(frozen=True)
class Calibration:
    valid_from: date
    v40: float
    v100: float


@dataclass(frozen=True)
class Zone:
    name: str
    moisture_voltage_measurement: str
    soil_temperature_measurement: str | None
    calibrations: tuple[Calibration, ...]


def _parse_zone(raw: dict) -> Zone:
    calibrations = tuple(
        sorted(
            (
                Calibration(
                    valid_from=date.fromisoformat(c["valid_from"]),
                    v40=float(c["v40"]),
                    v100=float(c["v100"]),
                )
                for c in raw["calibrations"]
            ),
            key=lambda c: c.valid_from,
        )
    )
    if not calibrations:
        raise ValueError(f"zone {raw['name']!r} without calibrations")
    for c in calibrations:
        if c.v100 <= 0 or c.v40 <= c.v100:
            raise ValueError(f"zone {raw['name']!r} has an invalid calibration: v40 must be greater than v100")
    return Zone(
        name=raw["name"],
        moisture_voltage_measurement=raw["moisture_voltage_measurement"],
        soil_temperature_measurement=raw.get("soil_temperature_measurement"),
        calibrations=calibrations,
    )


@lru_cache(maxsize=None)
def load_zones(path: str | None = None) -> tuple[Zone, ...]:
    zones_file = Path(path or os.getenv("ZONES_FILE") or DEFAULT_ZONES_FILE)
    with zones_file.open(encoding="utf-8") as f:
        raw = json.load(f)
    zones = tuple(_parse_zone(z) for z in raw.get("zones", []))
    if not zones:
        raise ValueError(f"no zones defined in {zones_file}")
    names = [z.name for z in zones]
    if len(names) != len(set(names)):
        raise ValueError(f"duplicated zones in {zones_file}")
    return zones


def zone_names() -> list[str]:
    return [z.name for z in load_zones()]


def voltages_to_percentage(df: pd.DataFrame, zone: Zone) -> pd.Series:
    times = pd.to_datetime(df["_time"], utc=True)
    result = pd.Series(float("nan"), index=df.index)
    bounds = [pd.Timestamp(c.valid_from, tz="UTC") for c in zone.calibrations] + [None]
    for calibration, start, end in zip(zone.calibrations, bounds, bounds[1:]):
        mask = times >= start
        if end is not None:
            mask &= times < end
        result[mask] = 100.0 - (df.loc[mask, "voltage"] - calibration.v100) * 60.0 / (
            calibration.v40 - calibration.v100
        )
    return result
