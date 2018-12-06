-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               10.1.22-MariaDB - mariadb.org binary distribution
-- Server OS:                    Win64
-- HeidiSQL Version:             9.4.0.5125
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;


-- Dumping database structure for coebot
CREATE DATABASE IF NOT EXISTS `coebot` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;
USE `coebot`;

-- Dumping structure for table coebot.aliases
CREATE TABLE IF NOT EXISTS `aliases` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `key` varchar(128) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `alias` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  KEY `FK_aliases_channel` (`channel_ID`),
  KEY `FK_aliases_commands` (`key`),
  CONSTRAINT `FK_aliases_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_aliases_commands` FOREIGN KEY (`key`) REFERENCES `commands` (`key`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.autoreplies
CREATE TABLE IF NOT EXISTS `autoreplies` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `trigger` varchar(512) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `response` varchar(512) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `index` bigint(20) unsigned NOT NULL,
  KEY `FK_autoreplies_channel` (`channel_ID`),
  KEY `created_at` (`index`),
  CONSTRAINT `FK_autoreplies_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.channel
CREATE TABLE IF NOT EXISTS `channel` (
  `channel_name` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '0',
  `enabled` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `commandPrefix` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '!',
  `bullet` varchar(32) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'coeBot',
  `cooldown` smallint(5) unsigned DEFAULT '5',
  `mode` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `lastfm` varchar(128) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `extraLifeID` varchar(16) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `timeoutDuration` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '600',
  `commercialLength` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '30',
  `signKicks` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `enableWarnings` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `useFilters` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `steamID` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `shouldModerate` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `subscriberAlert` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `subscriberRegulars` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `subMessage` varchar(256) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '(_1_) has subscribed!',
  `clickToTweetFormat` varchar(256) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Checkout (_CHANNEL_URL_) playing (_GAME_) on @TwitchTV',
  `parseYoutube` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `urbanEnabled` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `rollTimeout` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `rollLevel` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'regulars',
  `rollCooldown` smallint(5) unsigned NOT NULL DEFAULT '5',
  `rollDefault` smallint(5) unsigned NOT NULL DEFAULT '20',
  PRIMARY KEY (`channel_ID`),
  UNIQUE KEY `channel_name` (`channel_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- Data exporting was unselected.
-- Dumping structure for table coebot.commands
CREATE TABLE IF NOT EXISTS `commands` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `editor` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `count` int(10) unsigned DEFAULT NULL,
  `restriction` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `key` varchar(128) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `value` varchar(500) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `enabled` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  KEY `key` (`key`),
  KEY `FK_commands_channel` (`channel_ID`),
  CONSTRAINT `FK_commands_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.filters
CREATE TABLE IF NOT EXISTS `filters` (
  `channel_ID` varchar(32) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `filterLinks` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterOffensive` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterCaps` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterEmotesSingle` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterSymbols` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterEmotes` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterMe` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `filterSymbolsPercent` tinyint(4) NOT NULL DEFAULT '50',
  `filterCapsPercent` tinyint(4) NOT NULL DEFAULT '50',
  `filterCapsMinCapitals` tinyint(4) NOT NULL DEFAULT '6',
  `filterCapsMinCharacters` tinyint(4) NOT NULL DEFAULT '6',
  `filterSymbolsMin` tinyint(4) NOT NULL DEFAULT '5',
  `filterEmotesMax` tinyint(4) NOT NULL DEFAULT '5',
  `filterMaxLength` mediumint(8) unsigned NOT NULL DEFAULT '500',
  KEY `FK_filters_channel` (`channel_ID`),
  CONSTRAINT `FK_filters_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- Data exporting was unselected.
-- Dumping structure for table coebot.ignored_users
CREATE TABLE IF NOT EXISTS `ignored_users` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  KEY `channel_ID_regulars` (`channel_ID`),
  CONSTRAINT `FK_ignored_users_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.lists
CREATE TABLE IF NOT EXISTS `lists` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `list_name` varchar(128) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `restriction` tinyint(3) unsigned NOT NULL DEFAULT '1',
  KEY `channel_ID_regulars` (`channel_ID`),
  KEY `list_name` (`list_name`),
  CONSTRAINT `FK_lists_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.list_items
CREATE TABLE IF NOT EXISTS `list_items` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `list_name` varchar(128) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `item` varchar(500) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `index` bigint(20) unsigned NOT NULL,
  KEY `channel_ID_regulars` (`channel_ID`),
  KEY `FK_list_items_lists` (`list_name`),
  KEY `created_at` (`index`),
  CONSTRAINT `FK_list_items_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_list_items_lists` FOREIGN KEY (`list_name`) REFERENCES `lists` (`list_name`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.moderators
CREATE TABLE IF NOT EXISTS `moderators` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  KEY `FK_moderators_channel` (`channel_ID`),
  CONSTRAINT `FK_moderators_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- Data exporting was unselected.
-- Dumping structure for table coebot.offensivewords
CREATE TABLE IF NOT EXISTS `offensivewords` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `phrase` varchar(500) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  KEY `channel_ID_regulars` (`channel_ID`),
  CONSTRAINT `FK_offensivewords_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.owners
CREATE TABLE IF NOT EXISTS `owners` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  KEY `channel_ID_regulars` (`channel_ID`),
  CONSTRAINT `FK_owners_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.permitteddomains
CREATE TABLE IF NOT EXISTS `permitteddomains` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `domain` varchar(500) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  KEY `channel_ID_regulars` (`channel_ID`),
  CONSTRAINT `FK_permitteddomains_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.regulars
CREATE TABLE IF NOT EXISTS `regulars` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  KEY `FK_regulars_channel` (`channel_ID`),
  CONSTRAINT `FK_regulars_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- Data exporting was unselected.
-- Dumping structure for table coebot.repeated_commands
CREATE TABLE IF NOT EXISTS `repeated_commands` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `active` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `delay` smallint(5) unsigned NOT NULL DEFAULT '30',
  `messageDifference` smallint(5) unsigned NOT NULL DEFAULT '1',
  KEY `channel_ID_aliases` (`channel_ID`),
  KEY `FK_aliases_commands` (`name`),
  CONSTRAINT `FK_repeated_commands_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_repeated_commands_commands` FOREIGN KEY (`name`) REFERENCES `commands` (`key`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.scheduled_commands
CREATE TABLE IF NOT EXISTS `scheduled_commands` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `active` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `pattern` varchar(32) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `messageDifferrence` smallint(5) unsigned NOT NULL DEFAULT '1',
  KEY `channel_ID_aliases` (`channel_ID`),
  KEY `FK_aliases_commands` (`name`),
  CONSTRAINT `FK_scheduled_commands_channel` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_scheduled_commands_commands` FOREIGN KEY (`name`) REFERENCES `commands` (`key`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
-- Dumping structure for table coebot.users
CREATE TABLE IF NOT EXISTS `users` (
  `channel_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `user_ID` varchar(16) COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `permitted` varchar(1) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'N',
  `userlevel` enum('owner','moderator','regular','none','ignored') COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'none',
  `warning` bit(1) NOT NULL DEFAULT b'0',
  KEY `channel_ID_regulars` (`channel_ID`),
  CONSTRAINT `users_ibfk_1` FOREIGN KEY (`channel_ID`) REFERENCES `channel` (`channel_ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci ROW_FORMAT=COMPACT;

-- Data exporting was unselected.
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
