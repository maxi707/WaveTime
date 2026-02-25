INSERT INTO pools (name, address, timezone, is_active)
SELECT 'WaveTime Center', 'Москва, Ленинградский проспект, 15', 'Europe/Moscow', TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM pools WHERE name = 'WaveTime Center' AND address = 'Москва, Ленинградский проспект, 15'
);

INSERT INTO pools (name, address, timezone, is_active)
SELECT 'WaveTime South', 'Москва, Варшавское шоссе, 118', 'Europe/Moscow', TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM pools WHERE name = 'WaveTime South' AND address = 'Москва, Варшавское шоссе, 118'
);

INSERT INTO training_types (name, duration_minutes, default_capacity, price, is_active)
SELECT 'Свободное плавание', 60, 14, 800.00, TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM training_types WHERE name = 'Свободное плавание'
);

INSERT INTO training_types (name, duration_minutes, default_capacity, price, is_active)
SELECT 'Групповая тренировка', 60, 10, 1200.00, TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM training_types WHERE name = 'Групповая тренировка'
);
