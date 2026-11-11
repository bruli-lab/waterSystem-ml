import sys
from pathlib import Path

import numpy as np
import pandas as pd
import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import drying  # noqa: E402
import shadow  # noqa: E402

T0 = pd.Timestamp("2026-07-10 10:00", tz="UTC")


def readings(values_per_hour, start=T0, per_hour=6):
    rows = []
    for hour, value in enumerate(values_per_hour):
        for k in range(per_hour):
            rows.append({"_time": start + pd.Timedelta(hours=hour, minutes=10 * k), "soil_moisture": value})
    return pd.DataFrame(rows)


def no_waterings():
    return pd.DataFrame(columns=["_time", "seconds"])


def no_weather():
    return pd.DataFrame(columns=["_time", "temperature", "weather_is_raining_last"])


def no_soil_temperature():
    return pd.DataFrame(columns=["_time", "soil_temperature"])


def no_forecast():
    return pd.DataFrame(columns=["_time", "forecast_generated_at", *drying.FORECAST_COLUMNS])


def frame(values, waterings=None, weather=None, forecast=None):
    return drying.build_frame(
        drying.hourly_humidity(readings(values)),
        waterings if waterings is not None else no_waterings(),
        weather if weather is not None else no_weather(),
        no_soil_temperature(),
        forecast if forecast is not None else no_forecast(),
    )


def test_hourly_humidity_needs_three_readings():
    df = pd.concat([readings([60]), readings([58], start=T0 + pd.Timedelta(hours=1), per_hour=2)])
    hourly = drying.hourly_humidity(df)
    assert list(hourly.index) == [T0]


def test_hourly_humidity_drops_days_with_low_median():
    hourly = drying.hourly_humidity(readings([10, 10, 10]))
    assert hourly.empty


def test_hourly_humidity_aligned_to_end():
    end = T0 + pd.Timedelta(hours=2, minutes=5)
    hourly = drying.hourly_humidity(readings([60, 58, 56]), end=end)
    assert end - pd.Timedelta(hours=1) in hourly.index


def test_build_frame_deltas_and_rates():
    f = frame([60, 59, 57, 54, 50])
    assert f.loc[T0, "delta"] == pytest.approx(-1)
    assert f.loc[T0 + pd.Timedelta(hours=3), "delta_prev"] == pytest.approx(-3)
    assert f.loc[T0 + pd.Timedelta(hours=3), "rate3"] == pytest.approx(-2)
    assert f["valid"].sum() == 4


def test_build_frame_excludes_hours_near_watering_and_rises():
    waterings = pd.DataFrame({"_time": [T0 + pd.Timedelta(hours=2, minutes=30)], "seconds": [30]})
    f = frame([60, 59, 58, 90, 89, 88, 87, 86], waterings=waterings)
    near = [T0 + pd.Timedelta(hours=h) for h in (1, 2, 3)]
    assert not f.loc[near, "valid"].any()
    assert f.loc[T0 + pd.Timedelta(hours=5), "hours_since_watering"] == pytest.approx(3.5)
    assert f.loc[T0 + pd.Timedelta(hours=5), "last_seconds"] == 30

    rise = frame([60, 59, 70, 69])
    assert rise.loc[T0 + pd.Timedelta(hours=1), "rise_without_watering"]
    assert not rise.loc[T0 + pd.Timedelta(hours=1), "valid"]


def test_build_frame_excludes_rain():
    weather = pd.DataFrame({
        "_time": [T0 + pd.Timedelta(hours=3)], "temperature": [30.0], "weather_is_raining_last": [1],
    })
    f = frame([60, 59, 58, 57, 56, 55, 54, 53], weather=weather)
    assert not f.loc[[T0 + pd.Timedelta(hours=h) for h in (1, 2, 3, 4)], "valid"].any()
    assert f.loc[T0 + pd.Timedelta(hours=5), "valid"]


def test_build_frame_uses_only_forecast_known_at_decision_time():
    target = T0 + pd.Timedelta(hours=1)
    forecast = pd.DataFrame({
        "_time": [target, target],
        "forecast_generated_at": [T0 - pd.Timedelta(hours=5), T0 + pd.Timedelta(hours=3)],
        **{c: [1.0, 2.0] for c in drying.FORECAST_COLUMNS},
    })
    f = frame([60, 59, 58], forecast=forecast)
    assert f.loc[T0, "forecast_temperature"] == 1.0


def test_training_samples_counts_discarded_rises():
    X, y, stats = drying.training_samples(frame([60, 59, 70, 69, 68]))
    assert list(X.columns) == drying.FEATURES
    assert stats["discarded_rises"] == 1
    assert len(X) == len(y) == stats["samples"]


class ConstantDrying:
    def __init__(self, delta):
        self.delta = delta

    def predict(self, X):
        assert list(X.columns) == drying.FEATURES
        return np.full(len(X), self.delta)


def test_rollout_accumulates_deltas_without_forecast():
    f = frame([60, 59, 58])
    start_time = T0 + pd.Timedelta(hours=2)
    trajectory = drying.rollout(ConstantDrying(-2.0), f.loc[start_time], start_time, 4, no_forecast(), start_time)
    assert trajectory.tolist() == pytest.approx([56, 54, 52, 50])


@pytest.mark.parametrize(
    "local, steps, night",
    [
        ("2026-07-10 12:00", 3, False),
        ("2026-07-10 22:30", 11, False),
        ("2026-07-11 02:00", 7, True),
        ("2026-07-11 08:30", 1, True),
    ],
)
def test_horizon(local, steps, night):
    now = pd.Timestamp(local, tz=drying.LOCAL_TZ).tz_convert("UTC")
    assert shadow.horizon(now) == (steps, night)


def test_predict_zone_without_model(monkeypatch):
    monkeypatch.setattr(shadow, "drying_model_exists", lambda zone: False)
    result = shadow.predict_zone("Bonsai big", T0, no_weather(), no_forecast())
    assert result == {"zone": "Bonsai big", "error": "drying_model_not_found"}


@pytest.fixture
def live_data(monkeypatch):
    def setup(values, delta):
        now = T0 + pd.Timedelta(hours=len(values), minutes=-5)
        data = readings(values)
        monkeypatch.setattr(shadow, "drying_model_exists", lambda zone: True)
        monkeypatch.setattr(shadow, "load_drying_models",
                            lambda zone: (ConstantDrying(delta), ConstantDrying(delta), {"trained_at": "x"}))
        monkeypatch.setattr(shadow, "load_soil_moisture_data", lambda **kw: data.assign(zone=kw["zone"]))
        monkeypatch.setattr(shadow, "load_logs_data", lambda **kw: pd.DataFrame(columns=["_time", "seconds", "zone"]))
        monkeypatch.setattr(shadow, "load_soil_temperature_data", lambda **kw: no_soil_temperature())
        return now
    return setup


def test_predict_zone_would_water_when_trajectory_drops_below_40(live_data):
    now = live_data([50, 48, 46, 45], delta=-3.0)
    result = shadow.predict_zone("Bonsai big", now, no_weather(), no_forecast())
    assert result["horizon_hours"] == 3
    assert result["min_pessimistic"] == pytest.approx(36)
    assert result["would_water"] is True


def test_predict_zone_does_not_water_above_band(live_data):
    now = live_data([80, 79, 78], delta=-1.0)
    result = shadow.predict_zone("Bonsai big", now, no_weather(), no_forecast())
    assert result["would_water"] is False
    assert "error" not in result


def test_predict_zone_without_recent_readings(live_data):
    now = live_data([50, 48], delta=-1.0) + pd.Timedelta(hours=2)
    result = shadow.predict_zone("Bonsai big", now, no_weather(), no_forecast())
    assert result["error"] == "soil_measure_unavailable"
