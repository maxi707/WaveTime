DELETE FROM training_types
WHERE name IN ('Свободное плавание', 'Групповая тренировка');

DELETE FROM pools
WHERE (name = 'WaveTime Center' AND address = 'Москва, Ленинградский проспект, 15')
   OR (name = 'WaveTime South' AND address = 'Москва, Варшавское шоссе, 118');
