import argparse
import json
import os
import sys
from datetime import UTC, datetime

import joblib
import sklearn
from data_loader import (
    load_forecast_data,
    load_logs_data,
    load_soil_moisture_data,
    load_soil_temperature_data,
    load_weather_data,
)
from drying import FEATURES, build_frame, hourly_humidity, training_samples
from features import normalize_zone
from model_io import MODEL_DIR, slugify_zone
from sklearn.ensemble import HistGradientBoostingRegressor

TRAINING_START = "-365d"
PESSIMISTIC_QUANTILE = 0.2
MIN_SAMPLES = 200
MODEL_PARAMS = {"max_iter": 300, "learning_rate": 0.05, "max_leaf_nodes": 15, "min_samples_leaf": 30, "random_state": 0}


def drying_dir(zone: str) -> str:
    return os.path.join(MODEL_DIR, slugify_zone(zone))


def fit_models(X, y):
    median = HistGradientBoostingRegressor(loss="absolute_error", **MODEL_PARAMS).fit(X, y)
    pessimistic = HistGradientBoostingRegressor(
        loss="quantile", quantile=PESSIMISTIC_QUANTILE, **MODEL_PARAMS
    ).fit(X, y)
    return median, pessimistic


def save_drying_models(zone: str, median, pessimistic, metadata: dict) -> None:
    zone_dir = drying_dir(zone)
    os.makedirs(zone_dir, exist_ok=True)
    joblib.dump(median, os.path.join(zone_dir, "drying_median.joblib"))
    joblib.dump(pessimistic, os.path.join(zone_dir, "drying_p20.joblib"))
    with open(os.path.join(zone_dir, "drying_metadata.json"), "w", encoding="utf-8") as f:
        json.dump(metadata, f, indent=2)


def train_zone(zone: str) -> bool:
    zone = normalize_zone(zone)
    print(f"=== Entrenant model d'assecament: {zone} ===", file=sys.stderr)

    readings = load_soil_moisture_data(zone=zone, start=TRAINING_START, clip=False)
    waterings = load_logs_data(zone=zone, start=TRAINING_START)
    waterings = waterings[waterings["zone"] == zone]
    weather = load_weather_data(start=TRAINING_START)
    soil_temperature = load_soil_temperature_data(zone=zone, start=TRAINING_START)
    forecast = load_forecast_data(start=TRAINING_START)

    frame = build_frame(hourly_humidity(readings), waterings, weather, soil_temperature, forecast)
    X, y, stats = training_samples(frame)
    print(f"Mostres: {stats}", file=sys.stderr)

    if len(X) < MIN_SAMPLES:
        print(f"[WARN] {zone}: només {len(X)} mostres, no s'entrena", file=sys.stderr)
        return False

    median, pessimistic = fit_models(X, y)
    metadata = {
        "zone": zone,
        "trained_at": datetime.now(UTC).isoformat(),
        "features": FEATURES,
        "pessimistic_quantile": PESSIMISTIC_QUANTILE,
        "sklearn_version": sklearn.__version__,
        "data_from": X.index.min().isoformat(),
        "data_to": X.index.max().isoformat(),
        **stats,
    }
    save_drying_models(zone, median, pessimistic, metadata)
    print(f"Model d'assecament desat per a {zone}", file=sys.stderr)
    return True


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Entrena el model d'assecament d'una zona.")
    parser.add_argument("--zone", required=True, help='Zona a entrenar, per exemple: "Bonsai big".')
    parser.add_argument("--if-missing", action="store_true", help="Entrena només si la zona no té model.")
    return parser.parse_args()


def drying_model_exists(zone: str) -> bool:
    return os.path.exists(os.path.join(drying_dir(normalize_zone(zone)), "drying_metadata.json"))


def main() -> None:
    args = parse_args()
    if args.if_missing and drying_model_exists(args.zone):
        print(f"{args.zone}: ja té model d'assecament", file=sys.stderr)
        return
    if not train_zone(args.zone):
        sys.exit(1)


if __name__ == "__main__":
    main()
