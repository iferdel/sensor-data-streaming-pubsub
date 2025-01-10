SELECT time, measurement 
FROM sensor_measurement 
WHERE 
  time >= $__timeFrom()::timestamptz AND time < $__timeTo()::timestamptz 
ORDER BY time;
