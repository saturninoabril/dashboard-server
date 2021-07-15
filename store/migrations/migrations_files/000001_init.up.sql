CREATE TABLE IF NOT EXISTS {{.prefix}}System (
    Key    VARCHAR(64) PRIMARY KEY,
    Value  VARCHAR(1024) NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}Users (
    ID         CHAR(26) PRIMARY KEY,
    CreateAt   BIGINT NOT NULL,
    UpdateAt   BIGINT NOT NULL,
    Email      VARCHAR(64) NOT NULL UNIQUE,
    EmailVerified  BOOLEAN NOT NULL,
    Password   VARCHAR(128) NOT NULL,
    FirstName  VARCHAR(64) NOT NULL,
    LastName   VARCHAR(64) NOT NULL,
    State TEXT DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS {{.prefix}}Session (
    ID         CHAR(26) PRIMARY KEY,
    Token      CHAR(26) NOT NULL,
    CreateAt   BIGINT NOT NULL,
    ExpiresAt  BIGINT NOT NULL,
    UserID     VARCHAR(64) NOT NULL,
    CSRFToken  CHAR(26) NOT NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}Tokens (
    Token     VARCHAR(64) PRIMARY KEY,
    CreateAt  BIGINT DEFAULT NULL,
    Type      VARCHAR(64) DEFAULT NULL,
    Extra     VARCHAR(256) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}Roles (
    ID        CHAR(26) PRIMARY KEY,
    Name	  TEXT NOT NULL UNIQUE,
    CreateAt  BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    UpdateAt  BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)
);

CREATE TABLE IF NOT EXISTS {{.prefix}}UserRoles (
    UserID      CHAR(26) NOT NULL,
    RoleID		CHAR(26) NOT NULL,
    CreateAt    BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    UpdateAt    BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    PRIMARY     KEY (UserID, RoleID)
);
