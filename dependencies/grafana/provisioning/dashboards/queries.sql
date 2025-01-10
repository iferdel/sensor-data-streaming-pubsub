SELECT time, measurement 
FROM sensor_measurement 
WHERE 
  time >= $__timeFrom()::timestamptz AND time < $__timeTo()::timestamptz 
ORDER BY time;

SELECT time_bucket_gapfill('$bucket_interval', time) as time,
AVG(sensor_measurement.measurement) as measurement,
sensor.serial_number
FROM sensor_measurement INNER JOIN sensor ON sensor_measurement.sensor_id = sensor.id 
WHERE serial_number in ($serial_number)
AND time >= $__timeFrom()::timestamptz AND time < $__timeTo()::timestamptz
GROUP BY time_bucket_gapfill('$bucket_interval', time), serial_number
ORDER BY time;

SELECT sensor_measurement.time, sensor_measurement.measurement, sensor.serial_number FROM sensor_measurement INNER JOIN sensor ON sensor_measurement.sensor_id = sensor.id WHERE sensor.serial_number in ($serial_number) ORDER BY time;
