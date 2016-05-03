CREATE INDEX `iii2` ON `tech_db`.`thread` (`forumParentId` ASC);

CREATE INDEX `iii3` ON `tech_db`.`thread` (`userCreatedId` ASC);

CREATE INDEX `datethread` ON `tech_db`.`thread` (`dateU` ASC);

CREATE INDEX `ii2` ON `tech_db`.`post` (`postParentId` ASC);

CREATE INDEX `ii3` ON `tech_db`.`post` (`threadParentId` ASC);

CREATE INDEX `ii4` ON `tech_db`.`post` (`userCreatedId` ASC);

CREATE INDEX `ii5` ON `tech_db`.`post` (`forumParentId` ASC);

CREATE INDEX `iipath` ON `tech_db`.`post` (`postPath` ASC);

CREATE INDEX `levelidx` ON `tech_db`.`post` (`level` ASC);

CREATE INDEX `levelnumidx` ON `tech_db`.`post` (`levelnum` ASC);

CREATE INDEX `datepost` ON `tech_db`.`post` (`dateU` ASC);

CREATE INDEX `idx1` ON `tech_db`.`follow` (`ktoId` ASC);

CREATE INDEX `idx2` ON `tech_db`.`follow` (`kogoId` ASC);

CREATE INDEX `idxs1` ON `tech_db`.`subscribe` (`userId` ASC);
