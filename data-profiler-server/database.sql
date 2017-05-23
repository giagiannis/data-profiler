CREATE TABLE IF NOT EXISTS `datasets` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(200),
		`name` VARCHAR(200),
		`description` VARCHAR(2000)
);

CREATE TABLE IF NOT EXISTS `estimators` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`configuration` VARCHAR(2000),
		`datasetid` INTEGER,
		FOREIGN KEY(datasetid) REFERENCES datasets(id)
);

CREATE TABLE IF NOT EXISTS `matrices` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`configuration` VARCHAR(2000),
		`datasetid` INTEGER,
		FOREIGN KEY(datasetid) REFERENCES datasets(id)
);
