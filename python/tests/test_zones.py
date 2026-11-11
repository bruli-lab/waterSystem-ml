import json
import sys
from pathlib import Path

import pandas as pd
import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import data_loader  # noqa: E402
import zones  # noqa: E402


def write_zones(tmp_path, content):
    path = tmp_path / "zones.json"
    path.write_text(json.dumps(content), encoding="utf-8")
    return str(path)


def zone(name="a", calibrations=None):
    return {
        "name": name,
        "moisture_voltage_measurement": f"sensor.{name}_voltage",
        "soil_temperature_measurement": f"sensor.{name}_temperature",
        "calibrations": calibrations
        if calibrations is not None
        else [{"valid_from": "2026-04-01", "v40": 1.5, "v100": 1.3}],
    }


@pytest.fixture(autouse=True)
def clear_cache():
    zones.load_zones.cache_clear()
    yield
    zones.load_zones.cache_clear()


def test_manifest_file_loads():
    assert zones.zone_names() == ["Bonsai big", "Bonsai small"]


@pytest.mark.parametrize(
    "content",
    [
        {"zones": []},
        {"zones": [zone(calibrations=[])]},
        {"zones": [zone(calibrations=[{"valid_from": "2026-04-01", "v40": 1.2, "v100": 1.3}])]},
        {"zones": [zone(), zone()]},
    ],
)
def test_invalid_files(tmp_path, content):
    with pytest.raises(ValueError):
        zones.load_zones(write_zones(tmp_path, content))


def test_voltages_use_calibration_valid_at_each_time(tmp_path):
    z = zones.load_zones(
        write_zones(
            tmp_path,
            {
                "zones": [
                    zone(
                        calibrations=[
                            {"valid_from": "2026-09-01", "v40": 1.6, "v100": 1.4},
                            {"valid_from": "2026-04-01", "v40": 1.5, "v100": 1.3},
                        ]
                    )
                ]
            },
        )
    )[0]
    df = pd.DataFrame(
        {
            "_time": pd.to_datetime(["2026-03-31T12:00:00Z", "2026-06-01T12:00:00Z", "2026-10-01T12:00:00Z"]),
            "voltage": [1.3, 1.5, 1.5],
        }
    )
    result = zones.voltages_to_percentage(df, z)
    assert pd.isna(result.iloc[0])
    assert result.iloc[1] == pytest.approx(40.0)
    assert result.iloc[2] == pytest.approx(70.0)


def test_load_soil_moisture_converts_and_clamps(tmp_path, monkeypatch):
    monkeypatch.setenv("ZONES_FILE", write_zones(tmp_path, {"zones": [zone("a"), zone("b")]}))
    raw = pd.DataFrame(
        {
            "_time": ["2026-05-01T10:00:00Z", "2026-05-01T10:10:00Z", "2026-05-01T10:00:00Z"],
            "_measurement": ["sensor.a_voltage", "sensor.a_voltage", "sensor.b_voltage"],
            "_value": [1.2, 1.5, 1.8],
        }
    )
    monkeypatch.setattr(data_loader, "_query_to_df", lambda query, name: raw)

    df = data_loader.load_soil_moisture_data()

    assert list(df["zone"]) == ["a", "a", "b"]
    assert list(df["soil_moisture"]) == pytest.approx([100.0, 40.0, 0.0])

    only_b = data_loader.load_soil_moisture_data(zone="b")
    assert list(only_b["zone"]) == ["b"]
