CREATE TABLE irrigation_predictions (
                                        id UUID PRIMARY KEY,

                                        zone_id TEXT NOT NULL,              -- bonsai_big, bonsai_small...
                                        predicted_at TIMESTAMPTZ NOT NULL,

    -- Estat del sensor en el moment de decidir
                                        soil_moisture_percent NUMERIC(5,2),
                                        soil_voltage NUMERIC(6,3),
                                        root_temperature NUMERIC(5,2),
                                        terrace_temperature NUMERIC(5,2),
                                        is_raining BOOLEAN,

    -- Dades meteorològiques previstes/usades pel model
                                        forecast_temperature NUMERIC(5,2),
                                        forecast_humidity NUMERIC(5,2),
                                        forecast_precipitation NUMERIC(6,2),
                                        forecast_cloud_cover NUMERIC(5,2),
                                        forecast_shortwave NUMERIC(8,2),

    -- Context calculat
                                        days_since_last_watering NUMERIC(6,2),
                                        drying_factor NUMERIC(8,4),
                                        hour_of_day INT,
                                        month INT,

    -- Resultat del model
                                        model_version TEXT NOT NULL,
                                        should_water BOOLEAN NOT NULL,
                                        predicted_seconds INT,
                                        probability NUMERIC(6,4),

    -- Decisió final aplicada pel sistema
                                        decision_reason TEXT,               -- below_40, above_60, ml_prediction, night_blocked...
                                        executed BOOLEAN NOT NULL DEFAULT false,
                                        executed_seconds INT,

    -- Resultat posterior per validar
                                        moisture_before NUMERIC(5,2),
                                        moisture_after_30m NUMERIC(5,2),
                                        moisture_after_2h NUMERIC(5,2),
                                        moisture_after_24h NUMERIC(5,2),

                                        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);