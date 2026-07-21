UPDATE promotions
SET description = REPLACE(description, 'TSh ', 'Rp')
WHERE name IN ('Diskon Grand Opening', 'Promo Paket Hemat', 'Happy Hour 15%')
  AND description LIKE '%TSh %';
