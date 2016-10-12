CREATE TABLE `proto` (
  `createdAt` int(10) unsigned NOT NULL,
  `expiryAt` int(10) unsigned NOT NULL,
  `keyHash` int(10) unsigned NOT NULL,
  `key` varchar(1024) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (`createdAt`,`keyHash`),
  KEY `keyIdx` (`keyHash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED;
