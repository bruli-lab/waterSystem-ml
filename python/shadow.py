import json
import math
import os
import sys
import traceback

import numpy as np
import pandas as pd

from data_loader import (
    load_forecast_data,
    load_logs_data,
    load_soil_moisture_data,
    load_soil_temperature_data,
    load_weather_data,
)
from drying import HOUR, LOCAL_TZ, build_frame, hourly_humidity, rollout
from model_io import MODEL_DIR, _safe_joblib_load, slugify_zone
from zones import zone_names

DAY_HORIZON_HOURS = 3
NIGHT_START_HOUR = 23
NIGHT_END_HOUR = 9
LOW_THRESHOLD = 40.0
HIGH_THRESHOLD = 60.0
MAX_READING_AGE = pd.Timedelta(minutes=40)


def horizon(now: pd.Timestamp) -> tuple[int, bool]:
    local = now.tz_convert(LOCAL_TZ)
    night = local.hour >= NIGHT_START_HOUR or local.hour < NIGHT_END_HOUR
    if not night and local.hour < NIGHT_START_HOUR - 1:
        return DAY_HORIZON_HOURS, False
    end = local.normalize() + pd.Timedelta(hours=NIGHT_END_HOUR)
    if end <= local:
        end += pd.Timedelta(days=1)
    return max(1, math.ceil((end - local) / HOUR)), night


def load_drying_models(zone: str):
    zone_dir = os.path.join(MODEL_DIR, slugify_zone(zone))
    with open(os.path.join(zone_dir, "drying_metadata.json"), encoding="utf-8") as f:
        metadata = json.load(f)
    median = _safe_joblib_load(os.path.join(zone_dir, "drying_median.joblib"))
    pessimistic = _safe_joblib_load(os.path.join(zone_dir, "drying_p20.joblib"))
    return median, pessimistic, metadata


def drying_model_exists(zone: str) -> bool:
    return os.path.exists(os.path.join(MODEL_DIR, slugify_zone(zone), "drying_metadata.json"))


def _round(value) -> float | None:
    return None if value is None or pd.isna(value) else round(float(value), 2)


def predict_zone(zone: str, now: pd.Timestamp, weather: pd.DataFrame, forecast: pd.DataFrame) -> dict:
    if not drying_model_exists(zone):
        return {"zone": zone, "error": "drying_model_not_found"}
    median, pessimistic, metadata = load_drying_models(zone)

    readings = load_soil_moisture_data(zone=zone, start="-6h", clip=False)
    if readings.empty or now - readings["_time"].max() > MAX_READING_AGE:
        return {"zone": zone, "error": "soil_measure_unavailable"}
    humidity = float(np.clip(readings.sort_values("_time")["soil_moisture"].iloc[-1], 0.0, 100.0))

    hourly = hourly_humidity(readings, end=now)
    start_time = now - HOUR
    if start_time not in hourly.index:
        return {"zone": zone, "humidity": _round(humidity), "error": "not_enough_readings"}

    waterings = load_logs_data(zone=zone, start="-30d")
    waterings = waterings[waterings["zone"] == zone]
    soil_temperature = load_soil_temperature_data(zone=zone, start="-3h")
    frame = build_frame(hourly, waterings, weather, soil_temperature, forecast)
    start = frame.loc[start_time]

    steps, night = horizon(now)
    p20 = rollout(pessimistic, start, start_time, steps, forecast, now)
    med = rollout(median, start, start_time, steps, forecast, now)

    future = forecast[
        (pd.to_datetime(forecast["_time"], utc=True) >= now)
        & (pd.to_datetime(forecast["_time"], utc=True) < now + HOUR * steps)
    ] if not forecast.empty else forecast
    rain_probability = future["forecast_precipitation_probability"].max() if not future.empty else None

    min_pessimistic = float(p20.min())
    in_band = LOW_THRESHOLD <= humidity <= HIGH_THRESHOLD
    return {
        "zone": zone,
        "humidity": _round(humidity),
        "start_humidity": _round(start["h"]),
        "horizon_hours": steps,
        "night": night,
        "min_pessimistic": _round(min_pessimistic),
        "min_median": _round(med.min()),
        "pessimistic_1h": _round(p20[0]),
        "pessimistic_3h": _round(p20[min(2, steps - 1)]),
        "median_1h": _round(med[0]),
        "median_3h": _round(med[min(2, steps - 1)]),
        "rain_probability_max": _round(rain_probability),
        "would_water": bool(in_band and not night and min_pessimistic < LOW_THRESHOLD),
        "model_trained_at": metadata.get("trained_at"),
    }


def main() -> None:
    now = pd.Timestamp.now(tz="UTC")
    weather = load_weather_data(start="-6h")
    forecast = load_forecast_data(start="-2d", stop="2d")
    results = []
    for zone in zone_names():
        try:
            results.append(predict_zone(zone, now, weather, forecast))
        except Exception as e:
            traceback.print_exc(file=sys.stderr)
            results.append({"zone": zone, "error": str(e)})
    print(json.dumps(results))


if __name__ == "__main__":
    main()
