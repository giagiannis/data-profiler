CREATE TABLE IF NOT EXISTS `datasets` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(200),
		`name` VARCHAR(200),
		`description` VARCHAR(2000)
);

--CREATE TABLE IF NOT EXISTS `operators` (
--		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
--		`path` VARCHAR(200),
--		`name` VARCHAR(200),
--		`description` VARCHAR(2000)
--);
