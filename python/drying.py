import numpy as np
import pandas as pd

FORECAST_COLUMNS = [
    "forecast_temperature",
    "forecast_relative_humidity",
    "forecast_cloud_cover",
    "forecast_shortwave_radiation",
    "forecast_drying_factor",
]

FEATURES = [
    "h",
    "hour_sin",
    "hour_cos",
    "hours_since_watering",
    "last_seconds",
    "soil_temperature",
    "bme_temperature",
    *FORECAST_COLUMNS,
    "delta_prev",
    "rate3",
]

LOCAL_TZ = "Europe/Madrid"
MIN_READINGS_PER_HOUR = 3
LOW_DAY_MEDIAN = 15.0
MAX_RISE_WITHOUT_WATERING = 5.0
WATERING_BEFORE = pd.Timedelta(hours=1)
WATERING_AFTER = pd.Timedelta(hours=2)
RAIN_HOURS = (-1, 0, 1, 2)
HOUR = pd.Timedelta(hours=1)


def hourly_humidity(readings: pd.DataFrame, end: pd.Timestamp | None = None) -> pd.Series:
    if readings.empty:
        return pd.Series(dtype=float)
    df = readings[["_time", "soil_moisture"]].dropna().copy()
    df["_time"] = pd.to_datetime(df["_time"], utc=True)
    day = df["_time"].dt.floor("D")
    median = df.groupby(day)["soil_moisture"].transform("median")
    df = df[median >= LOW_DAY_MEDIAN]
    if df.empty:
        return pd.Series(dtype=float)
    if end is None:
        start = df["_time"].dt.floor("h")
    else:
        start = end - HOUR * np.ceil((end - df["_time"]) / HOUR)
    grouped = df.groupby(start)["soil_moisture"].agg(["mean", "count"])
    grouped = grouped[grouped["count"] >= MIN_READINGS_PER_HOUR]
    series = grouped["mean"]
    series.index.name = "_time"
    return series.sort_index()


def _forecast_for(times: pd.DatetimeIndex, known_at: pd.DatetimeIndex, forecast: pd.DataFrame) -> pd.DataFrame:
    empty = pd.DataFrame(np.nan, index=range(len(times)), columns=FORECAST_COLUMNS)
    if forecast.empty or len(times) == 0:
        return empty
    fc = forecast.dropna(subset=["forecast_generated_at"]).copy()
    if fc.empty:
        return empty
    fc["target"] = pd.to_datetime(fc["_time"], utc=True).dt.floor("h")
    fc = fc.sort_values("forecast_generated_at")
    left = pd.DataFrame({"target": times.floor("h"), "known_at": known_at, "row": range(len(times))})
    merged = pd.merge_asof(
        left.sort_values("known_at"),
        fc[["target", "forecast_generated_at", *FORECAST_COLUMNS]],
        left_on="known_at",
        right_on="forecast_generated_at",
        by="target",
        direction="backward",
    )
    return merged.sort_values("row").reset_index(drop=True)[FORECAST_COLUMNS]


def _hour_features(times: pd.DatetimeIndex) -> tuple[np.ndarray, np.ndarray]:
    local = (times + HOUR / 2).tz_convert(LOCAL_TZ)
    hour = local.hour + local.minute / 60
    return np.sin(2 * np.pi * hour / 24), np.cos(2 * np.pi * hour / 24)


def _rain_near(times: pd.DatetimeIndex, rain_hours: set) -> np.ndarray:
    floored = times.floor("h")
    return np.array([any((t + HOUR * k) in rain_hours for k in RAIN_HOURS) for t in floored], dtype=bool)


def build_frame(
    hourly: pd.Series,
    waterings: pd.DataFrame,
    weather: pd.DataFrame,
    soil_temperature: pd.DataFrame,
    forecast: pd.DataFrame,
) -> pd.DataFrame:
    if hourly.empty:
        return pd.DataFrame(columns=["h", "delta", "valid", "rise_without_watering", *FEATURES])
    idx = pd.DatetimeIndex(hourly.index)
    df = pd.DataFrame({"h": hourly.values}, index=idx)

    w = waterings.dropna(subset=["_time"]).sort_values("_time")
    w_ns = _ns(pd.DatetimeIndex(pd.to_datetime(w["_time"], utc=True)))
    lo = np.searchsorted(w_ns, _ns(idx - WATERING_BEFORE))
    hi = np.searchsorted(w_ns, _ns(idx + WATERING_AFTER))
    watering_near = hi > lo

    rain_hours = set()
    if not weather.empty:
        rain = weather[weather["weather_is_raining_last"] == 1]
        rain_hours = set(pd.to_datetime(rain["_time"], utc=True).dt.floor("h"))
    rain_near = _rain_near(idx, rain_hours)

    nxt = hourly.reindex(idx + HOUR).values
    df["delta"] = nxt - df["h"].values
    clean = ~watering_near & ~rain_near & ~np.isnan(nxt)
    df["rise_without_watering"] = clean & (df["delta"].values > MAX_RISE_WITHOUT_WATERING)
    df["valid"] = clean & ~df["rise_without_watering"]

    delta_ok = df["delta"].where(df["valid"])
    prev = [delta_ok.reindex(idx - HOUR * k).values for k in (1, 2, 3)]
    df["delta_prev"] = prev[0]
    stacked = np.vstack(prev)
    counts = np.sum(~np.isnan(stacked), axis=0)
    sums = np.nansum(stacked, axis=0)
    df["rate3"] = np.where(counts > 0, sums / np.maximum(counts, 1), np.nan)

    decided_at = idx + HOUR
    last_pos = np.searchsorted(w_ns, _ns(decided_at), side="right") - 1
    has_last = last_pos >= 0
    safe_pos = np.clip(last_pos, 0, None)
    if len(w_ns):
        elapsed = (_ns(decided_at) - w_ns[safe_pos]) / 3.6e12
        seconds = pd.to_numeric(w["seconds"], errors="coerce").to_numpy(dtype=float)[safe_pos]
    else:
        elapsed = seconds = np.full(len(idx), np.nan)
    df["hours_since_watering"] = np.where(has_last, elapsed, np.nan)
    df["last_seconds"] = np.where(has_last, seconds, np.nan)

    df["soil_temperature"] = _asof(soil_temperature, "soil_temperature", decided_at)
    df["bme_temperature"] = _asof(weather, "temperature", decided_at)

    fc = _forecast_for(idx + HOUR, decided_at, forecast)
    for c in FORECAST_COLUMNS:
        df[c] = fc[c].values

    df["hour_sin"], df["hour_cos"] = _hour_features(idx)
    return df


def _ns(times: pd.DatetimeIndex) -> np.ndarray:
    return pd.DatetimeIndex(times).tz_convert("UTC").as_unit("ns").asi8


def _asof(df: pd.DataFrame, column: str, times: pd.DatetimeIndex) -> np.ndarray:
    if df.empty or column not in df.columns:
        return np.full(len(times), np.nan)
    series = df.assign(_time=pd.to_datetime(df["_time"], utc=True)).dropna(subset=[column]).sort_values("_time")
    if series.empty:
        return np.full(len(times), np.nan)
    pos = np.searchsorted(_ns(pd.DatetimeIndex(series["_time"])), _ns(times), side="right") - 1
    values = pd.to_numeric(series[column], errors="coerce").to_numpy(dtype=float)
    return np.where(pos >= 0, values[np.clip(pos, 0, None)], np.nan)


def training_samples(frame: pd.DataFrame) -> tuple[pd.DataFrame, pd.Series, dict]:
    rises = int(frame["rise_without_watering"].sum())
    samples = frame[frame["valid"]]
    stats = {
        "hours": len(frame),
        "samples": len(samples),
        "discarded_rises": rises,
    }
    return samples[FEATURES], samples["delta"], stats


def rollout(model, start: pd.Series, start_time: pd.Timestamp, steps: int, forecast: pd.DataFrame,
            known_at: pd.Timestamp) -> np.ndarray:
    times = pd.DatetimeIndex([start_time + HOUR * k for k in range(steps)])
    fc = _forecast_for(times + HOUR, pd.DatetimeIndex([known_at] * steps), forecast)
    sin, cos = _hour_features(times)
    h = float(start["h"])
    deltas = [start["delta_prev"]] if not pd.isna(start["delta_prev"]) else []
    rate3 = start["rate3"]
    trajectory = []
    for k in range(steps):
        row = start[FEATURES].astype(float).copy()
        row["h"] = h
        row["hour_sin"], row["hour_cos"] = sin[k], cos[k]
        row["hours_since_watering"] = start["hours_since_watering"] + k
        for c in FORECAST_COLUMNS:
            row[c] = fc[c].iloc[k]
        if k > 0:
            row["delta_prev"] = deltas[-1]
            row["rate3"] = float(np.mean(deltas[-3:])) if deltas else rate3
        delta = float(model.predict(pd.DataFrame([row[FEATURES]]))[0])
        deltas.append(delta)
        h += delta
        trajectory.append(h)
    return np.array(trajectory)
